tinycat
=======

A minuscule MARC search system. See it in action: https://asciinema.org/a/14073

Build an index from a MARC file. This will only use the first `245.a` field for now.

    $ tinycat -input file.mrc

Query:

    $ tinycat -q Berlin
