# since the tag plugin edits files which need to be generated by other plugins first this dedicated file exists
version: v1
managed:
  enabled: true
plugins:  
  - name: gotag
    out: ./
    opt: outdir={{ (ds "data").relativeGoLibOutDir }},auto=bun-as-snake