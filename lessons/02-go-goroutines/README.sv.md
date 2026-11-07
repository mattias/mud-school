# Lektion 02 – Många spelare samtidigt

Testade du att ansluta två gånger samtidigt i förra lektionen? Den andra
spelaren fick inte ens välkomstmeddelandet förrän den första kopplade ner.
Servern gör bara en sak i taget. En MUD utan andra spelare är ganska ensam,
så nu fixar vi det.

## Mål

- Förstå vad en **goroutine** är och starta en med `go`.
- Förstå skillnaden mellan *samtidighet* (concurrency) och *parallellism*.
- Använda `sync.WaitGroup` för att vänta in flera goroutines.
- Köra testerna med *race detector*: `./mudschool check --race`.

## Goroutines

En goroutine är en lättviktig tråd som Gos runtime schemalägger. De är
billiga – en server kan ha hundratusentals. Du startar en genom att sätta
`go` framför ett funktionsanrop:

```go
go s.handle(conn)   // returnerar direkt; handle körs "vid sidan av"
```

Anonyma funktioner (closures) fungerar också:

```go
for i := 0; i < 3; i++ {
    go func() {
        fmt.Println("hej från goroutine", i)
    }()
}
```

> Sedan Go 1.22 får varje varv i en `for`-loop en egen kopia av `i`. I äldre
> Go-versioner var det en klassisk bugg att alla goroutines såg samma `i`.

Vill du vänta tills ett gäng goroutines är klara använder du en `WaitGroup`:

```go
var wg sync.WaitGroup
for _, job := range jobs {
    wg.Add(1)
    go func() {
        defer wg.Done()
        do(job)
    }()
}
wg.Wait()
```

Titta i `mudschool_02_test.go` – testerna använder exakt det mönstret för
att låta 50 klienter skicka samtidigt.

## Uppgift

1. Ändra `Serve` så att varje anslutning hanteras i en egen goroutine.
2. Logga när en klient ansluter och kopplar ner med `log.Printf`
   (`conn.RemoteAddr()` ger klientens adress). Testerna kontrollerar inte
   loggarna, men du kommer att vilja ha dem.

Det är en ändring på ett tecken plus loggning – men begreppet är det
viktigaste i hela Go-delen av kursen.

## Testa själv

Starta servern och anslut från tre terminaler samtidigt (`./mudschool connect`). Kör sedan
`./mudschool check --race`. Race detectorn hittar ställen där två goroutines
rör samma data utan synkronisering. Än så länge delar dina goroutines ingen
data – men vänta bara till lektion 05!

## Fundera på

- Vad händer om tusen klienter ansluter? Vad kostar en goroutine i minne?
  (Svar: några kilobyte stack som växer vid behov.)
- `handle` blockerar på `scanner.Scan()` medan den väntar på klienten.
  Varför är det helt okej nu, men inte i lektion 01?

## Utmaningar

- Räkna antal anslutna klienter och logga det: lägg till ett vanligt fält
  `clients int` i `Server`, räkna upp i början av `handle` och ner (med
  `defer`) när klienten går. Kör `./mudschool check --race`: du bör få
  `WARNING: DATA RACE`, eftersom många `handle`-goroutiner ändrar samma
  variabel samtidigt. Läs om `sync/atomic` (`atomic.Int64`) och fixa det.
  (Räknar du i stället i `Serve`-loopen försvinner varningen – men varför? Och
  var skulle du räkna *ner* då? Race-detektorn rapporterar bara race som
  faktiskt inträffar medan testerna kör.)
