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