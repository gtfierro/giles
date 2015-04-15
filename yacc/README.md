## sMAP Query Lang Rewrite

Using Go yacc!

This command is just for testing the where clause parsing

```bash
$ go tool yacc -o main.go -p SQ where.y && rlwrap go run main.go
```
