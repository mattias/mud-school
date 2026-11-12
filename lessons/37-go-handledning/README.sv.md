# Lektion 37 – Introduktionen

Fas 6 tar ut din MUD i världen, och där ute är de flesta spelare nya. Någon
som aldrig har spelat en MUD loggar in, ser en rumsbeskrivning och en
blinkande markör – och går. Botemedlet är lika gammalt som MUD:arna: en
**introduktion**. Nya spelare börjar i ett eget litet område och leds steg
för steg:

```
Begynnelsens sal
En stilla sal med vitkalkade väggar. ...
[Introduktion 1/6] Välkommen! Den här korta introduktionen lär dig grunderna
(tutorial skip lämnar den). Skriv look för att se var du är.
> look
...
[Introduktion 2/6] Det finns en dörröppning norrut. Gå igenom den med: go north
> go north
Gården
...
[Introduktion 3/6] Något ligger på marken här. Plocka upp det med: get fackla
```

När sista steget är klart kliver de ut i den riktiga världen.

## Mål

- Beskriv beteende som **data**: stegen finns i världsfilen, inte i koden.
- Återanvänd händelsesystemet från lektion 34 till något nytt: en hook som
  ser vad spelare *gör*, i stället för att stoppa dem.
- Ändra en databas som redan används: en ny **migrering**.
- Hitta rummen man kan nå från ett rum: en **bredden-först-sökning**.

## Go-koncept

### Steg som data

```json
"tutorial": {
  "start": "tutorial-hall",
  "steps": [
    {"text": "Skriv look för att se var du är.", "command": "look"},
    {"text": "Gå norrut.", "room": "tutorial-yard"},
    {"text": "Plocka upp den med: get fackla", "item": "tutorial-torch"}
  ],
  "done": "Bra gjort! ..."
}
```

Varje steg säger *vad spelaren ska få höra* och *när det är klart* – efter
ett kommando, i ett rum eller med en sak i packningen. Servern vet ingenting
om facklor eller gårdar. Att lägga till ett steg, översätta ett eller bygga
en helt ny introduktion är en ändring i en fil, inte i koden.

### En hook som tittar

Händelserna från lektion 34 kommer *innan* något händer, så att hooks kan
säga nej. Introduktionen behöver det omvända: att få veta vad som just har
hänt. En ny händelse, `EventCommand`, kommer efter varje kommando som finns:

```go
s.emit(&Event{Kind: EventCommand, Player: p, Room: p.Room, Text: cmd.Verb})
```

och introduktionen är bara en hook: `s.On(EventCommand, s.onTutorialCommand)`.
Skript får den på köpet: `mud.on("command", ...)`.

### En migrering

Lektion 20 lovade att databasen kan växa utan att kastas bort:

```go
// 4: how far they have come in the tutorial (lesson 37); 0 = not in it
`ALTER TABLE players ADD COLUMN tutorial INTEGER NOT NULL DEFAULT 0`,
```

`DEFAULT 0` är det som gör det säkert: varje spelare som sparades innan
introduktionen fanns får 0 – "inte i introduktionen" – och skickas aldrig
tillbaka till skolan.

### Introduktionens område

Introduktionens rum måste vara avskilda: en ny spelare som kan gå ut ur
introduktionen har hoppat över den av misstag. Området är alla rum man kan
nå från starten – en bredden-först-sökning över utgångarna, med en kö och en
`map[string]bool` med rum man redan sett. Samma område talar om för servern
vilka saker som ska läggas tillbaka när en spelare går, så att nästa nykomling
också hittar facklan.

## Kontraktet

- En världsfil får ha ett `"tutorial"`-objekt: `start` (ett rum), `steps`
  (vart och ett med `text` och precis ett av `command`, `room`, `item`) och
  `done`. `Tutorial`, `TutorialStep`, `World.Tutorial`; `worldFile.Tutorial`.
- `LoadWorld` kontrollerar introduktionen efter sakerna och rapporterar alla
  fel på en gång: `MsgErrTutorialStart` (och sedan inget annat),
  `MsgErrTutorialOpen` (världens start går att nå från introduktionen),
  `MsgErrTutorialEmpty`, och per steg (numrerat från 1) `MsgErrTutorialStep`,
  `MsgErrTutorialRoom` (inte i introduktionens område), `MsgErrTutorialItem`
  (ligger inte i det).
- `(w *World) LoadTutorial(r)` lägger till en introduktionsfil – `rooms`,
  `items` och `tutorial` – i en värld; `DefaultWorld` laddar `tutorial.json` (given) med den.
- En spelare utan sparning (eller utan databas) börjar i introduktionens
  startrum, på steg 1, och får `MsgTutorialStep` (nummer, antal, text) efter
  rumsbeskrivningen.
- `EventCommand` kommer efter varje kommando som finns (inbyggt, skript eller
  plugin, inte ett okänt), med verbet – alias expanderade – i `Text`.
- Efter varje kommando passeras steg som är klara, flera på en gång om det
  behövs; nästa steg visas. Efter det sista: `done`-texten, saker från
  introduktionens område läggs tillbaka i sina rum, och spelaren flyttas till
  världens startrum (`MsgDisappears`, `MsgArrives`, en look).
- `tutorial` upprepar steget; `tutorial skip` lämnar introduktionen
  (`MsgTutorialSkipped`); utanför den: `MsgTutorialNotIn`.
- `SavedPlayer.Tutorial` (steget, 0 = inte i introduktionen) sparas i en ny
  kolumn genom migrering 4; en spelare som loggar in igen fortsätter där hen var
  (på sista steget om introduktionen har blivit kortare). En spelare som *inte*
  är i introduktionen men står i dess område flyttas till världens start – hen
  skulle aldrig komma ut. En värld som laddas om utan introduktion skickar
  alla i den till starten (`moveToStart`, given).
- En copyover (lektion 35) behåller steget: `CopyoverPlayer.Tutorial`.
- `worldSnapshot` (lektion 36) behåller introduktionen.

## Uppgift

1. `tutorial.go` är delvis given (typerna, `tutorialFile`, `moveToStart`): skriv
   `LoadTutorial`, `tutorialArea`, `validateTutorial`, `startTutorial`,
   `showTutorialStep`, `stepDone`, `onTutorialCommand`, `leaveTutorial` och
   `cmdTutorial`.
2. `world.go`: `Tutorial` på `World` och `worldFile`, och kontrollen i `LoadWorld`.
3. `events.go` och `dispatch`: `EventCommand`. `server.go`: prenumerera med
   hooken, låt nya spelare börja i introduktionen, visa steget efter första looken.
4. `player.go`: `tutorial int`. `store.go`: migrering 4, `SavedPlayer.Tutorial`
   och kolumnen i `SavePlayer`/`LoadPlayer`. `account.go`: spara och återställ den.
5. `defaultworld.go`: bädda in `tutorial.json` och `LoadTutorial` den.
   `commands.go`: `tutorial`. `olc.go`: `Tutorial` i `worldSnapshot`.
   `copyover.go`: `Tutorial` i `CopyoverPlayer`, `PrepareCopyover` och `adopt`.

Logga sedan in med ett nytt namn och gör introduktionen själv.

## Fundera på

- Två nya spelare kommer samtidigt, och det finns en fackla. Vad händer? Hur
  löser riktiga MUD:ar det? (Sök på "instancing".)
- Varför kommer hooken *efter* kommandot? Vad skulle gå fel om introduktionen
  använde de avbrytbara händelserna från lektion 34?
- `DEFAULT 0` gör att gamla spelare hoppar över introduktionen. När skulle
  du vilja ha det omvända?
- Vad mer i din MUD skulle kunna vara data i stället för kod?

## Utmaningar

- Ett steg som är klart när spelaren *säger* något särskilt, eller besegrar
  en varelse (`EventDeath`).
- En introduktion i Lua: låt skript lägga till steg med `mud.on("command", ...)`.
- Tips: om en spelare inte har klarat ett steg på en minut, upprepa det med andra ord.
