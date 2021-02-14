# immudbcc

The Immudb code challenge: Design and implement a document-like data model based upon ImmuDB.

# 1. Problem Statement

A document data model is composed of identifiable elements which have properties and may be related to other elements in
many ways. For more information about document data models, we've listed some helpful resources.

We expect a minimal functional implementation written in Go (including unit testing) - either fully integrated into
Immudb or as a standalone component. The new immutable document data model needs to be built on top of the Immudb immutable
key-value model in order to leverage immutability properties.

In developing your solution, you'll have to make decisions that are related to:

1. API which stores and retrieves documents and provide proof of no tampering.
2. Data format which defines how objects are exchanged between final application and storage.
3. Atomicity at the level of document property, single, or multiple document.
4. Concurrency to allow concurrent read and update operations.
5. Granularity of evidence to prove whether a whole document or only the document property is unchanged and untempered.

---

# 2. Solution Proposal

## 2.1 Object Insertion.

The main solution revolves around the observation that to properly leverage ImmuDB natural features, consists in designing
a solution capable of marshalling generic JSON objects, and storing each object's properties as a Key-Value entries.

Considering that a JSON object has a tree-like structure, it follows, that any Key can be described as the full path from
the object's root until that specific leaf, that has a specific value.

To illustrate this, let's consider the following example:

```
object = {
            "name":"John",
            "age":30,
            "cars": {
                "car1":"Ford",
                "car2":"BMW",
                "car3":"Fiat"
            },
            "available": true
          }
```

So, ideally our API would be able to store an arbitrary JSON object associated to a chosed ID. For instance:

```
    Store("objectID", object)
```

This would trigger the transformation of the object, into a representation capable of efficiently be stored in ImmuDB.

* Step #1 - Marshaling the raw JSON document into and internal Key-Value pair format:

This JSON object, with an implicit `objectID` could be represented by the following set of key-value pairs:

```
    "objectID/cars/car1/string" = "Ford"
    "objectID/cars/car2/string" = "BMW"
    "objectID/cars/car3/string" = "Fiat"
    "objectID/name/string"      = "John"
    "objectID/age/float64"      = 30
    "objectID/available/bool"   = true
```

In the previous example, it is defined eah `Key` as the path from the root of the object up until the leaf, plus the
`type` of the `Value` being inserted. This latter meta-data will be useful during the unmarshaling phase, avoiding a
type assertion.

Note, that the `objectID` is used as a key prefix. This ID can be arbitrary, and decided by the application using this API,
especially given the fact that JSON objects do not follow a rigid structure, for instance, where an `id` field is not
guaranteed.

* Step #2 - Object database insertion:

Now, let's assume that the previous Key-Values are inserted in ImmuDB, and for which insertion we get a correspondent
insertion `index`.

```
    Set("objectID/cars/car1/string", "Ford") -> (idx_1, hash_1)
    Set("objectID/cars/car2/string","BMW")   -> (idx_2, hash_2)
    Set("objectID/cars/car3/string","Fiat")  -> (idx_3, hash_3)
    Set("objectID/name/string", "John")      -> (idx_4, hash_4)
    Set("objectID/age/float64", 30)          -> (idx_5, hash_5)
    Set("objectID/available/bool", true)     -> (idx_6, hash_6)
```

The order on how these properties are inserted in the database, is irrelevant, and therefore can be done concurrently.

* Step #3 - Object database commit manifest:

Moreover, this JSON object will only be consider committed / inserted when a special `object manifest` key is inserted
after the properties' insertion. The object manifest for this specific example would have the following format:

```
   manifest = {
                id: "<objectID>",
                indexes: [idx_1, idx_2, idx_3, idx_4, idx_5, idx_6],
                hash:  sha256(hash_1, hash_2, hash_3, hash_4, hash_5, hash_6)
              }

   Set("manifest/<objectID>", manifest.Marshall()) -> (gIdx, gHash)
```

The API would return as confirmation of the commit, the `index` of the object commit manifest, plus a `global hash`.

---

So, let's frame the problem more formally.

For a given JSON document object:
* Perform the marshaling of the object's tree:
    * Each path represents the tree's depth-first transversal.
    * Each path is prefixed by the assigned ObjectID.
    * Each path is suffixed by the basic type of the associated Value.
      (According to JSON, `string`, `float64`, `bool` and `null` are supported).
    * The end result is a set of Key-Value pairs, where the Key is the associated path aforementioned, and the Value is
    the property associated value.

* The previous set X of Key-Value pairs, can be concurrently stored in the database:
    * For a given `n-th` Key-Value, insert it in the database, storing in memory the resulting `n-th` Index and Hash.
        ```
        indexHashPair  = [ (index_0,hash_0), (index_1,hash_1), ..., (index_n,hash_n) ]
        ```
    * Perform the insertions for each element of X.
    * Sort the `indexHashPair` list, using the `index` as a `sorting criteria`. This ensures that the insertion order
    (given that it is concurrent) has no influence in subsequent calculations (i.e. global hash).
* Finally, when all the properties are inserted, we can then mark this insertion as successful by inserting the manifest:
    * The manifest is essentially a JSON payload containing the object's ID, the list of property indexes that compose
    it, and the properties' global hash.
        ```
        {
           "id":      "objectID",
           "hash":    indexHashPair.Hash(),    // GlobalHash(hash_0,  hash_1,  ..., hash_n)
           "indexes": indexHashPair.Indexes()  // List=[index_a, index_b, ..., index_n]
        }
        ```
    * Store the manifest in the database:
        ```
        Set("manifest/objectID", manifest.Marshal())
        ```
    * Return the final `index` and the `hash` of the object's manifest. These values in combination with the objectID are
    the minimal elements required to retrieve the object.

This use case was implemented in such a way it can be done concurrently, which is guaranteed by ImmuDB: all document 
properties and object manifests are written in parallel.

---

## 2.2 Object Retrieval:

This essentially leverages the commit manifest acting as a document summary in order to fetch the document associated 
properties. In essence, it performs the following actions:

* Step #1 - Read database object commit manifest:
  
  Fetches the `object manifest` entry associated with a specific `document ID`. From this, all property indexes of the 
  document's underlying properties are known, as well as the document's global Hash.
  
* Step #2 - Fetch all the document's properties:
  
  Given that at this point, all the document's property insertion indexes are known, it is now possible to fetch them 
  all, which is done in parallel. 

* Step #3 - Unmarshal the list of internal Key-Value pairs into the original raw JSON document:

  At this point, given that all the document's sub-properties are loaded, now it is simply required to call the custom 
  unmarshaller to retrieve the raw JSON document.
  
Finally, this use case was implemented in such a way it can be done concurrently, which is guaranteed by ImmuDB: all
document properties are fetched in parallel.

---

## 2.3 Object Update:

Given the nature of the applications that require a immutable database, it is fair to assume that updates will be not
very frequent, and far between. Therefore, the proposed API only considers single property updates.

Also, given the non-transactional nature of ImmuDB, it is assumed that the caller ofn this API will guarantee that a given
document is not getting updated concurrently. If that happens, the multiple callers with end-up with a fragmented result 
given that each resulting object manifest after the update will point to different sets of underlying Key-Values.


# 3. How to test and build the project.

To execute the linters and unit tests:
```
./run_ci.sh
```

To execute integration tests:
```
./run_itests.sh
```

To execute the previous commands in a Docker:
```
docker build -t immudb-cc -f Dockerfile .
```
