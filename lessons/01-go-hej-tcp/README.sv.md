# Lektion 01 – Hej, TCP!

Alla MUD:ar sedan 1978 börjar på samma sätt: en server som lyssnar på en
port och en spelare som ansluter med `telnet`. I den här lektionen bygger du
den allra första versionen av din server – en **ekoserver** som hälsar och
upprepar allt du skriver.

## Mål

- Förstå hur ett Go-program är uppbyggt: `package`, `import`, `func`.
- Använda en `struct` med en **metod**.
- Hantera fel på Go-sättet: `if err != nil { return err }`.
- Lyssna på TCP, ta emot en anslutning och läsa rader.

## Go på fem minuter

```go
package mud                 // varje fil tillhör ett paket

import "fmt"                // importera standardbiblioteket

type Server struct{}        // en struct (en "post" med fält – än så länge inga)

func NewServer() *Server {  // en funktion som returnerar en pekare till en Server
    return &Server{}
}

func (s *Server) Hello() string {  // en METOD på *Server – (s *Server) är "mottagaren"
    return "hej"
}

func divide(a, b int) (int, error) { // funktioner kan returnera flera värden
    if b == 0 {
        return 0, fmt.Errorf("kan inte dela %d med noll", a)
    }
    return a / b, nil               // nil betyder "inget fel"
}
```

Go har inga undantag (exceptions). Fel är vanliga värden av typen `error`
som du returnerar och kontrollerar. Det känns pratigt i början, men gör det
väldigt tydligt var saker kan gå fel.

`defer` schemalägger ett anrop som körs när funktionen returnerar – perfekt
för att städa upp:

```go
func (s *Server) handle(conn net.Conn) {
    defer conn.Close()   // körs oavsett hur vi lämnar funktionen
    ...
}
```

## Nätverk i Go

Att öppna porten är `main.go`s jobb. Den skapar *lyssnaren* och ger den till
din server:

```go
// main.go – redan skriven, du behöver inte ändra den
l, err := net.Listen("tcp", "127.0.0.1:4000")
srv.Serve(l)
```

`Serve` *använder* bara lyssnaren den får:

```go
conn, err := l.Accept()              // vänta på nästa klient – blockerar!
fmt.Fprint(conn, "hej\r\n")          // conn är en io.Writer ...
scanner := bufio.NewScanner(conn)    // ... och en io.Reader
for scanner.Scan() {                 // läs rad för rad
    line := scanner.Text()           // raden utan radbrytning
}
```

> **Anropa inte `net.Listen` inne i `Serve`.** Testerna ger `Serve` en
> lyssnare på en slumpad ledig port och ansluter till *den*. Öppnar `Serve` en
> egen port når testerna aldrig din server – och en andra lyssnare på samma
> port misslyckas, ofta med ett förvirrande `nil pointer dereference`.

> **Varför `\r\n`?** Telnet-protokollet (RFC 854) säger att en radbrytning är
> CR LF. Många klienter klarar `\n`, men vi gör rätt från början.

## Uppgift

Öppna `workspace/server/mud/server.go` och fyll i de två `TODO`:

1. `Serve` ska i en evig `for`-loop anropa `l.Accept()`. Får du ett fel ska
   `Serve` returnera det (det händer t.ex. när lyssnaren stängs). Annars
   anropar du `s.handle(conn)`.
2. `handle` ska stänga anslutningen med `defer`, skicka `Welcome` + `"\r\n"`
   och sedan eka varje rad den läser.

> **Om `messages.go`:** all text som spelet visar för spelaren ligger i
> `workspace/server/mud/messages.go` – på det språk du valt för kursen.
> `Welcome` finns redan där. Testerna använder samma konstanter, så behåll
> namnen, men ändra gärna texterna. Kod och kodkommentarer är på engelska,
> som är brukligt i branschen; instruktionerna är på ditt språk.

Kör `./mudschool check` för att testa.

## Testa själv

```
./mudschool run server            # i en terminal
telnet 127.0.0.1 4000             # i en annan (eller: nc / ncat 127.0.0.1 4000)
```

Skriv några rader. Öppna sedan en *andra* telnet samtidigt. Vad händer?
Varför? (Det fixar vi i nästa lektion.)

## Fundera på

- Vad händer om du skriver `l, err := ...` *inne i* `for`-loopen? (`:=`
  deklarerar en **ny** variabel som döljer parametern `l` – det kallas
  *shadowing*. `go vet` varnar inte för det som standard.)
- Varför tar `Serve` lyssnaren som parameter istället för att öppna porten
  själv? (Tips: tänk på testerna – och på kommande lektioner.)

## Utmaningar (frivilliga)

- Logga varje ny anslutning med `log.Printf("ny anslutning från %s", conn.RemoteAddr())`.
- Låt servern svara `MsgGoodbye` och koppla ner om klienten skriver `quit`.
- Läs dokumentationen: `go doc net.Listener` och `go doc bufio.Scanner`.
