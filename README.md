GalaxyZooChallenge
==================

Central pixel benchmark code for Kaggle's Galaxy Zoo competition.

This benchmark queries a 10x10 pixel patch in the center of each training set galaxy image. RGB values are averaged for these pixels and then hashed to a single value. This is done over all training images and the hash is designed to create clusters of like-colored centers.

For each cluster of colors, the 37 probability values (Class values) for each galaxy in the cluster are averaged.

For each test set image, find the color of the central patch, hash it. Find the matching cluster in the training set and assign the Class values for that cluster to the test galaxy.
