# xxd

Les inn en binærfil og gjør om heksadesimal.

## Hvordan kjøre programmet:

- `zig build run -- file1.txt` for å teste mot `file1.txt`
- Output blir:

```text
00000000: fffe 5400 6500 7300 7400 6900 6e00 6700  
00000010: 2000 6d00 6500 6400 2000 6d00 6500 7200
00000020: 2000 6f00 7500 7400 7000 7500 7400 0d00
00000030: 0a00
```