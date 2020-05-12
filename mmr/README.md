# merkle mountain range

Merkle Mountain Ranges (MMR) decomposes an arbitory number of leaves into a seiries of perfectly balanced merkle trees, MMR is strictly append-only, elements are added from the left to the right, adding a parent as soon as 2 children exist, filling up the range accordingly.

This illustrates a range with 11 inserted leaves and total size 19, where each node is annotated with its order of insertion:

```
Height

3              14
             /    \
            /      \
           /        \
          /          \
2        6            13
       /   \        /    \
1     2     5      9     12     17
     / \   / \    / \   /  \   /  \
0   0   1 3   4  7   8 10  11 15  16 18
```

## find inclusion proof visually

Suppose the path from root to the leaf `tn` is `(root, t1, t2, ... tn)`, then the inclusion proof is chosen by replacing `ti` with it's sibling. 

```
               hash
              /    \
             /      \
            /        \
           /          \
          /            \
         k              l
        / \            / \
       /   \          /   \
      /     \        /     \
     g       h      i      j
    / \     / \    / \     
    a b     c d    e f     

```

The path from root to `a` is `(hash, k, g, a)`, so the inclusion proof is `(b, h, l)`.

## find consistency proof visually

The consistency proof first has to prove the existence of the old root, and then prove the inclusion of the old root in the new root. So the first step is to identify the old root, then take the peak `P` of rightest merkle montain, then find the inclusion proof of `P` visually; If `P` equals to the old root(old tree size is power of two), the consistency proof is just `InclusionProof(P)`; If `P` doesn't equal to the old root(old tree size is not power of two), the consistency proof is `P || InclusionProof(P)`.

Using the method, it's not hard to see that `ConsistencyProot(c)` is `[c,d,g,i]`, `ConsistencyProot(d)` is `[l]`, `ConsistencyProot(f)` is `[i, j, k]`, 
