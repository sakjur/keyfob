# Keyfob

**Please note, Keyfob is still very much alpha-grade software.**

Keyfob is a medium-scale† per-user encryption key-management system.
For every user, Keyfob may track multiple namespaces which can be used
to keep track of different kinds of data (e.g. to separate the different
legal grouds for processing data as defined in the
EU General Data Protection Regulation). In addition to this, Keyfob uses
key-derivation where the root key is paired with a service key (presumably
unique to the service) which effectively gives every user a unique key for
every combination of namespace and service.

The design is heavily influenced by the [Scalable User Privacy](https://labs.spotify.com/2018/09/18/scalable-user-privacy/)
blog post from Spotify describing their Padlock service. One goal of Keyfob
is to provide the functionality described by Padlock while implementing some
of the improvements suggested in the blog post.

†: Keyfob is built to support high availability and to scale, but supporting
applications with millions of requests per second is not an explicit goal
of Keyfob.

## Build

TBD

## Documentation

- Keyfob was inspired by the blog post [Scalable User Privacy by Spotify Labs](https://labs.spotify.com/2018/09/18/scalable-user-privacy/).

## Copyright

This project is licensed under the Apache 2 license, which can be read in its
entirety in the LICENSE-file.

- Copyright 2019 Emil Tullstedt
