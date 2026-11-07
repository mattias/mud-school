# Lektion 27 – Vägval

Vargen i den mörka skogen har vandrat planlöst sedan lektion 16. Nu
**jagar** den: när den ser dig kommer den – runt hörn, förbi träd, den
kortaste vägen. Och råttan, svårt skadad, **flyr**.

```
###############        ###############        ###############
#.............#        #.............#        #.............#
.@.....w......#   ->   .@.w..........#   ->   .@w...........#   "En varg anfaller dig!"
#.............#        #.............#        #.............#
```

## Mål

- Implementera **A\***, den mest använda vägsökningsalgoritmen i spel.
- Skriva en **generisk** funktion med en typparameter begränsad av `comparable`.
- Använda `container/heap` – en prioritetskö – genom dess interface.
- Lära dig **Dijkstra-kartan**, schweizerkniven inom roguelike-AI.

## Go-koncept

### `container/heap`

Standardbiblioteket har ingen färdig prioritetskö – men ett paket som gör
*din* slice till en, om den implementerar `heap.Interface`:

```go
type queue[N any] struct{ items []*entry[N] }
func (q *queue[N]) Len() int           { return len(q.items) }
func (q *queue[N]) Less(i, j int) bool { return q.items[i].f < q.items[j].f }
func (q *queue[N]) Swap(i, j int)      { ... }
func (q *queue[N]) Push(x any)         { ... }  // anropas av heap.Push
func (q *queue[N]) Pop() any           { ... }  // anropas av heap.Pop: ta bort det SISTA elementet
```

Sedan håller `heap.Push(q, e)` och `heap.Pop(q)` det minsta `f` först – på
O(log n). Obs: du anropar `heap.Push`, inte `q.Push`.

### Generics med `comparable`

```go
func AStar[N comparable](start, goal N, neighbors func(N) []N,
	cost func(a, b N) int, h func(N) int) ([]N, bool)
```

`N` kan vara vad som helst som kan jämföras med `==` och användas som nyckel
i en map: `Pos` i en zon, `string` för rums-ID:n, `int` i ett test. Samma A*
hittar vägar i zoner *och* i världsgrafen.

## Algoritmerna

### A\*

Utforska den mest lovande noden först: den med lägst `f = g + h`, där `g` är
kostnaden hittills och `h` en *uppskattning* av det som återstår. Med steg i 8
riktningar som alla kostar 1 är den perfekta uppskattningen
**Chebyshev-avståndet**, `max(|dx|, |dy|)` – det överskattar aldrig, så vägen
som hittas är den kortaste.

```
open := en prioritetskö med start
så länge kön inte är tom:
	cur := noden med lägst f
	om cur == goal: gå bakåt genom came[...] – det är vägen
	för varje granne n:
		g := kostnaden hittills + cost(cur, n)
		om n är ny eller g är billigare än förut: kom ihåg g och came[n] = cur, lägg n i kön
```

### Dijkstra-kartor

En bredden-först-sökning från **alla mål på en gång** ger varje ruta dess
avstånd till närmaste mål. Sedan blir besluten triviala:

- gå mot målet: steg till grannen med *lägst* tal;
- kom undan: steg till grannen med *högst* tal.

*Brogue* använder Dijkstra-kartor för nästan allt monsterbeteende.

## Kontraktet

- `AStar[N comparable](start, goal, neighbors, cost, h) ([]N, bool)` – den
  billigaste vägen, **utan** `start` och **med** `goal`; `nil, true` när start
  är målet; `nil, false` när det inte finns någon väg.
- `(z *Zone) Path(from, to Pos, blocked func(Pos) bool) ([]Pos, bool)` – i en
  zon, 8 riktningar, kostnad 1, bara gångbara rutor, inte genom `blocked`-rutor
  (får vara `nil`) – utom `to`.
- `DijkstraMap(z, goals []Pos) map[Pos]int` – stegen till närmaste mål
  (8 riktningar) för varje nåbar gångbar ruta.
- Nya beteenden i NPC-filerna (`range` och `below` i `BehaviorSpec`):
  - `{"type": "hunt", "range": 5, "chance": 0.3}` – ser spelare med
    `ComputeFOV(zonen, sin position, range)` (standard 6); bredvid en spelare
    (Chebyshev ≤ 1) anfaller den: `prey.target = n`, `MsgAttacksYou` till
    bytet och `MsgAttacks` till rummet; annars tar den första steget av `Path`
    mot den närmaste synliga spelaren, inte genom andra varelser. `chance` =
    sannolikheten per tick att agera (0 = varje tick).
  - `{"type": "flee", "below": 0.4}` – under `below` × sin hälsa, med spelare
    i rummet: steg till grannen med högre `DijkstraMap`-värde än där den står.
- Standardvärlden: vargen jagar, råttan flyr.

## Uppgift

`path.go`: `AStar` (kön är given), `Path`, `DijkstraMap`, `Hunter.Act` och
`Fleer.Act`. Sedan `newBehavior` och `BehaviorSpec`. Prova: gå in i den mörka
skogen (`morkskogen`) och vänta.

## Fundera på

- Varför får `h` aldrig överskatta? Konstruera ett fall där ett för stort `h`
  ger en väg som inte är den kortaste.
- Vargen räknar ut en hel A* varje gång den tar ett steg. Med 100 vargar?
  Hur skulle en Dijkstra-karta från spelarna kunna ersätta dem alla?
- `AStar` fungerar på rum också. Skriv `w.RoomPath(from, to string)` – vad
  skulle `h` vara? (Tips: rumskoordinaterna från lektion 9.)
- Vad händer om två vargar vill till samma ruta?

## Utmaningar

- `travel <rum>`: gå automatiskt, ett steg per tick, till ett rum långt bort
  (A* i världsgrafen, sedan A* i varje zon).
- Flockar: vargar som försöker omringa dig (var och en väljer en egen ledig ruta bredvid dig).
- Flyende varelser som lämnar rummet när de når en öppning.
