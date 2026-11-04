# Idéer att bygga vidare på

> [🇬🇧 English](../en/IDEAS.md) · 🇸🇪 Svenska · [🇯🇵 日本語](../ja/IDEAS.md)

En samling idéer: till spelet, till klienten och till själva kursen. Många
passar som utmaningar, egna projekt eller framtida lektioner.

## Spelet

**Inspirerat av Dwarf Fortress och Caves of Qud**

- **Historiegenerering:** generera byns historia innan spelet startar –
  grundare, krig, försvunna artefakter – och låt NPC:er berätta om den.
  Böcker i världen som innehåller genererad historia.
- **Kroppsdelar istället för HP:** en träff i armen gör att du tappar svärdet.
- **Mutationer och fraktioner** (Qud): rykte hos olika grupper som påverkar hur
  varelser reagerar på dig.
- **Material och temperatur:** trä brinner, vatten fryser på vintern, lava
  smälter sten. Ett simuleringssystem som körs i spelloopen.
- **Ekologi:** vargar jagar rådjur, rådjur äter gräs. Om du jagar bort alla
  rådjur blir vargarna hungriga ...

**Klassiskt MUD**

- Dygn och väder: natt gör kartan mörkare, regn ger ljud och sämre sikt.
- Uppdrag (quests) skrivna i Lua (fas 5), med tillstånd per spelare.
- Hantverk: recept i datafiler, smedjan i Mudby blir användbar.
- Ekonomi: butiker, priser som påverkas av utbud och efterfrågan.
- Husdjur och följeslagare som använder vägsökningen från fas 4.
- Brev och anslagstavla som sparas mellan omstarter.
- Kanaler (`ooc`, `newbie`) och ignorera-lista.
- Grupper: dela erfarenhet och se varandra på kartan.

**Tekniska spelfunktioner**

- **Repriser:** eftersom protokollet är JSON-rader kan klienten spela in en
  session till fil och spela upp den igen – helt utan server.
- **Åskådarläge:** följ en annan spelare och se det hen ser.
- **Bottar:** skriv AI-spelare i Rust som utforskar världen (bra för lasttest
  och för att hitta buggar).
- **Prestationer och topplistor.**

## Klienten

- Minikarta över hela utforskade världen bredvid närkartan.
- Teman: valbara färgpaletter, inklusive **färgblindvänliga** (undvik att
  bara skilja grönt och rött åt) och en monokrom "fosforskärm"-variant.
- Skärmläsarläge: bara text, inga ramar – tillgänglighet är en viktig
  anledning till att MUD:ar fortfarande spelas.
- Makron och alias i klienten (`alias k kill`), triggers (`när texten "Du är
  hungrig" syns, skicka "eat bread"`).
- Klickbara utgångar och namn i loggen.
- Ljud: en enkel ljudmotor (t.ex. `rodio`) som spelar effekter från `sound`-frames.
- Tiles: stöd för flera tilesets som kan bytas med en tangent, inklusive
  "isometriskt" läge som utmaning.

## Kursen

- **Buggjakt-lektioner:** en lektion där kursverktyget *planterar* en bugg
  (en data race, en deadlock, ett läckande goroutine) som du ska hitta med
  `-race`, `pprof` eller loggning.
- **Refaktoreringslektioner:** "byt mutex mot actor" (fas 3) är en sådan –
  testerna är oförändrade, bara koden ändras.
- **Läs riktig kod:** korta guidade turer i källkoden till t.ex. `net/http`,
  `bufio` eller `ratatui` när de används.
- **Parprogrammering:** en person bygger servern, en klienten, och
  protokolldokumentet är det enda de delar. Byt roller varje fas.
- **Kodgranskning:** en checklista för att granska varandras lösningar.
- **Kata:** små fristående övningar per språk (t.ex. "implementera
  `ParseCommand` i Rust") för den som vill öva mer på ett koncept.
- **Webbversion av kursverktyget:** visa lektioner och testresultat i en
  webbläsare (Go: `net/http` + `html/template`), med möjlighet att köra i
  t.ex. GitHub Codespaces utan installation.
- **Prestationsmärken** i `mudschool list` för gjorda utmaningar.
- **Spelet på flera språk samtidigt:** varje spelare väljer språk vid inloggning
  och servern väljer `messages` per spelare. Bra övning i att designa om när
  globala konstanter inte längre räcker.
- **Fler språk:** samma protokoll gör det lätt att lägga till ett spår där
  klienten skrivs i ett tredje språk.
