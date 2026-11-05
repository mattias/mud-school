# Ledtrådar

## En evig loop

Go har bara en loop-sats: `for`. En evig loop skrivs `for { ... }`.

```go
for {
    conn, err := l.Accept()
    if err != nil {
        return err
    }
    // ...
}
```

## Använd lyssnaren du får

`Serve` får en lyssnare `l` som redan lyssnar. Anropa inte `net.Listen` i
`Serve`, och skriv inte `l, err := ...` i loopen – det skapar ett nytt `l` som
döljer det du fick. Skriv `conn, err := l.Accept()`.

## Skriva till anslutningen

`net.Conn` uppfyller gränssnittet `io.Writer`, så du kan använda
`fmt.Fprint(conn, "text\r\n")` eller `conn.Write([]byte("text\r\n"))`.
Glöm inte att importera `"fmt"` och `"bufio"` överst i filen.

## Läsa rader

```go
scanner := bufio.NewScanner(conn)
for scanner.Scan() {
    line := scanner.Text()
    fmt.Fprint(conn, line+"\r\n")
}
```

`scanner.Scan()` returnerar `false` när klienten kopplar ner – då tar loopen
slut, funktionen returnerar och din `defer conn.Close()` körs.
