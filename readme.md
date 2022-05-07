# Reptar

Reptar or Reproducible Tar creates tar archives from files that are stripped of information that typically makes reproducability challenging.

Generally this means that this library tries to implement something similar to this tar command.

```
tar - \
    --sort=name \
    --mtime="1970-01-01 00:00:00Z" \
    --owner=0 --group=0 --numeric-owner \
    --pax-option=exthdr.name=%d/PaxHeaders/%f,delete=atime,delete=ctime \
    -cf
```

Further reading here for more information on reproducible archives: https://reproducible-builds.org/docs/archives/
