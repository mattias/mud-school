# Lektion 26 – Synfält

Hittills får varje spelare hela zonen – också det som ligger bakom väggar,
runt hörn och i nästa rum. I ett spel är det fusk som väntar på att hända:
en modifierad klient skulle visa allt. Därför bestämmer servern vad varje
spelare **ser** och skickar bara det.

```
#####################      o  du
#*******************#      *  synligt
#*******************#      .  inte synligt (pelarens skugga)
#*******************#
#*******************#
#*********o*#.......#  <- pelaren (#) syns, det som är bakom den gör inte det
#*******************#
```

Två sorters kunskap:

- **synligt** – det du ser *just nu* (det ändras vid varje steg);
- **utforskat** – det du har sett *någon gång*. De rutorna ligger kvar på
  kartan, så du minns rummet du kom ifrån – "fog of war", krigets dimma.

## Mål

- Implementera **rekursiv skuggkastning** (shadowcasting), den klassiska roguelike-algoritmen.
- Använda rekursion, lutningar som flyttal och symmetri (en oktant, åtta gånger).
- Skilja på vad servern *vet* och vad en klient får *veta*.

## Algoritmen

Dela upp synfältet i åtta oktanter (trianglar). I varje skannar du rad för
rad, utåt. Håll reda på *lutningarna* `start` och `end` för den del av raden
som fortfarande är upplyst (från 1.0 till 0.0 till att börja med). När
skanningen möter en vägg:

1. anropa dig själv rekursivt för raderna bakom, men bara fram till väggens vänsterkant,
2. fortsätt efter väggen med en ny `start` – väggens högerkant.

När en rad slutar blockerad är oktanten klar: allt bakom ligger i skugga.

```go
for j := row; j <= radius; j++ {
	dy := -j
	for dx := -j; dx <= 0; dx++ {
		left := (float64(dx) - 0.5) / (float64(dy) + 0.5)
		right := (float64(dx) + 0.5) / (float64(dy) - 0.5)
		// hoppa över det som är utanför start..end, markera rutan, hantera väggar
	}
}
```

De åtta oktanterna är samma kod med koordinaterna speglade och ombytta:
`(dx, dy)` i oktanten blir `(dx*xx + dy*xy, dx*yx + dy*yy)` – tabellen
`octants` är given.

## Kontraktet

- `ComputeFOV(z *Zone, origin Pos, radius int) map[Pos]bool` – utgångspunkten
  är synlig; rutor inom `radius` (euklidiskt: `dx²+dy² <= radius²`) som en
  linje från utgångspunkten når; väggar, träd och berg syns men skymmer det
  bakom (`Transparent`, lektion 23); vatten skymmer inte.
- `PlayerSight = 8`.
- `updateFOV(p)` – sparar `p.fov`, lägger till det i `p.explored[rumsID]` och
  returnerar rutorna som var nya (i läsordning: rad för rad, vänster till höger).
- **Zone-frames:** `rows` innehåller bara utforskade rutor (resten är `' '`),
  förklaringen bara de tecken som förekommer, och `visible` är synfältet som
  rader av `'1'` och `'0'`. `entities`: bara det som syns – och alltid du själv.
  Telnet ser samma rader.
- **Delta-frames** får `tiles` – `[{x, y, code, meaning}]` för rutor som syns
  för första gången – och `visible` (alla rader) när synfältet har ändrats.
  Entiteter som försvinner ur sikte är `removed`, de som kommer in i sikte
  `added`. En delta skickas om *något* ändrats, även bara `visible`.
- Det utforskade minns per rum så länge du är inloggad.

## Uppgift

1. `fov.go`: `castLight` och `updateFOV` (`ComputeFOV`, `octants` och
   `visibleRows` är givna).
2. `Player`: `fov`, `explored`, `seenVisible`.
3. `sync.go`: `ZoneData.Visible`, `TileChange`, `DeltaData.Tiles` och
   `DeltaData.Visible`; låt `zoneData` visa bara de utforskade rutorna och
   `visibleEntities`; låt `syncPlayer` lägga till nya rutor och ändrat `visible`.

Din Rust-klient från lektion 25 ignorerar de nya fälten (serde ignorerar
okända fält) – den visar nyupptäckta rutor från lektion 28.

## Fundera på

- Varför måste synfältet räknas ut på servern och inte i klienten?
- `ComputeFOV` körs för varje spelare efter varje kommando och varje tick.
  Hur dyrt är det för 31×13? För 200×200 med 100 spelare? Vad skulle du kunna cacha?
- Är skuggkastning *symmetrisk* – om A ser B, ser B då A? Prova några fall.
  Varför spelar det roll i ett spel med avståndsattacker?
- Varelser har inget synfält än. I lektion 27 kommer de att behöva ett.

## Utmaningar

- Ljus: mörka grottor där synfältet är 3 om du inte bär en fackla.
- Låt dörrar (`+`) skymma sikten när de är stängda.
- Spara `explored` i databasen, så att kartan minns mellan besöken.
