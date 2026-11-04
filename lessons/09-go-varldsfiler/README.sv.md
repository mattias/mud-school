# Lektion 09 – Världsfiler

Att skriva rum i Go-kod fungerar för sex rum, men inte för sexhundra. Och
det går inte att ändra världen utan att kompilera om. Riktiga MUD:ar har
byggare (*builders*) som skapar områden i datafiler. Det är också första
steget mot en server som kan **ladda om innehåll medan den kör** (fas 5).

Kursverktyget har lagt in `workspace/server/mud/world.json` med 21 rum:
Mudby, landsvägen, sjön, skogen och en grotta. Varje rum har nu också
**terräng** och **koordinater** – det behöver vi för kartan i nästa lektion.

## Mål

- Läsa JSON med `encoding/json` och styra fältnamn med **struct-taggar**.
- Programmera mot gränssnitt: `io.Reader` istället för filnamn.
- Validera data ordentligt och rapportera *alla* fel med `errors.Join`.
- Baka in filer i programmet med `//go:embed`.

## Go-koncept

### JSON och struct-taggar

```go
type Room struct {
    ID    string            `json:"id"`
    Exits map[string]string `json:"exits,omitempty"`
}

var r Room
err := json.Unmarshal(data, &r)        // från []byte

dec := json.NewDecoder(reader)         // från vilken io.Reader som helst
dec.DisallowUnknownFields()            // "nmae" istället för "name" -> fel
err = dec.Decode(&r)
```

### io.Reader

`LoadWorld(r io.Reader)` kan läsa från en fil (`os.Open`), en sträng
(`strings.NewReader` – så gör testerna), en byte-slice (`bytes.NewReader`),
ett nätverk ... Gränssnitt i Go uppfylls *implicit*: alla typer med metoden
`Read([]byte) (int, error)` är en `io.Reader`.

### errors.Join

```go
var errs []error
errs = append(errs, fmt.Errorf("rum %q: ...", id))
return errors.Join(errs...)   // nil om listan är tom
```

En byggare som har gjort fem fel vill se alla fem på en gång.

### go:embed

```go
import _ "embed"

//go:embed world.json
var defaultWorldJSON []byte   // filens innehåll, inbakat vid kompilering
```

## Koordinatsystemet

```
        y-1 (north)
 x-1 (west) · x+1 (east)        z+1 = up, z-1 = down
        y+1 (south)
```

## Uppgift

1. **`world.go`**:
   - lägg till fälten `Terrain string`, `X`, `Y`, `Z int` i `Room` och
     JSON-taggar på alla fält (titta i `world.json` för namnen),
   - `LoadWorld(r io.Reader) (*World, error)` – läs formatet
     `{"start": "...", "rooms": [...]}`, med `DisallowUnknownFields`,
     ge fel om samma ID förekommer två gånger, och validera,
   - `LoadWorldFile(path string) (*World, error)`,
   - `(w *World) Validate() error` som hittar **alla** dessa fel:
     startrummet saknas · rum utan id eller namn · okänd terräng (inte i
     `Terrains`, se nedan) · okänd riktning · utgång till ett rum som inte
     finns · två rum med samma koordinater · en utgång vars koordinater inte
     stämmer (går man `north` måste målet ligga på `y-1`). Felmeddelandena
     finns färdiga i `messages.go` (`MsgErrStartMissing`, `MsgErrExitMissing`, ...);
     använd dem med `fmt.Errorf`,
   - `var Terrains = []string{"town", "inside", "road", "field", "forest", "water", "mountain", "cave"}`,
2. **`defaultworld.go`**: skriv om `DefaultWorld()` så att den läser den
   inbakade `world.json` (och får göra `panic` om den är trasig – det är ett
   programmeringsfel).
3. **`main.go`**: lägg till en flagga `-world fil.json` som laddar en annan värld.

## Testa själv

```
cd workspace/server
cp mud/world.json /tmp/min-varld.json   # ändra något, gör gärna fel med flit
go run . -world /tmp/min-varld.json
```

## Fundera på

- Varför är `[3]int` en giltig map-nyckel men inte `[]int`?
- Varför får `DefaultWorld` göra `panic` men inte `LoadWorldFile`?

## Utmaningar

- Bygg ett eget område på minst fem rum och lägg till det i `world.json`.
- Varna (logga, men ge inte fel) för enkelriktade utgångar.
- Skriv ett litet verktyg `go run ./cmd/worldcheck fil.json` som bara validerar.
