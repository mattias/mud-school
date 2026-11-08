# Lektion 32 – Skript i Lua

Att ladda om data (lektion 31) ändrar *vad* världen innehåller. För att ändra
*vad som händer* – ett nytt kommando, en varelse som beter sig annorlunda –
har du behövt skriva Go och starta om. Nu får servern ett **inbäddat
skriptspråk**: kommandon och beteenden i `.lua`-filer, omladdade i samma
ögonblick som du sparar.

```lua
-- scripts/dance.lua
mud.command("dance", function(player, args)
  mud.send(player, player .. " dances " .. args .. "!")
end)

mud.behavior("bark", function(npc, room)
  if mud.roll(1, 20) == 1 then mud.broadcast(room, "The dog barks.") end
end)
```

[Lua](https://www.lua.org) är *det* klassiska språket för att skripta spel
(World of Warcraft, Roblox, Factorio, …). Vi bäddar in det med
**gopher-lua** – en Lua-tolk skriven i ren Go, så ingen C-kompilator behövs.

## Mål

- Bädda in en tolk och ge den ett API i Go (`mud.send`, `mud.player`, …).
- **Sandlåda** för kod man inte litar på: inga filer, inga program, en tidsgräns.
- Använda `context.Context` för att stoppa kod som kör för länge.
- Tänka på API-design: vad ska ett skript kunna göra – och inte?

## Go-koncept

### Anropa Go från Lua

En Go-funktion med signaturen `func(L *lua.LState) int` kan anropas från
Lua: den läser sina argument från Lua-stacken (`L.CheckString(1)`), lägger
sina resultat på den (`L.Push(...)`) och returnerar hur många.

```go
"send": func(L *lua.LState) int {
	if p := s.players[NormalizeName(L.CheckString(1))]; p != nil {
		p.SendKind(KindInfo, L.CheckString(2))
	}
	return 0
},
```

`L.SetFuncs(tabell, map)` lägger en hel map av dem i en tabell, och
`L.SetGlobal("mud", tabell)` gör den till det globala `mud`.

### Sandlådan

`lua.NewState(lua.Options{SkipOpenLibs: true})` börjar med *ingenting*.
Öppna sedan bara det som är säkert – `base`, `table`, `string`, `math` – och
ta bort de funktioner i `base` som läser filer eller laddar kod (`dofile`,
`loadfile`, `load`, `loadstring`, `require`). Ingen `os`, ingen `io`: ett
skript kan inte radera filer eller starta program.

### En tidsgräns

```go
ctx, cancel := context.WithTimeout(context.Background(), ScriptTimeout)
defer cancel()
L.SetContext(ctx)
defer L.RemoveContext()
err := L.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}, args...)
```

Med en context kontrollerar gopher-lua mellan instruktionerna om tiden är
slut och stoppar skriptet med ett fel – även `while true do end`.
`Protect: true` gör Lua-fel till ett Go-`error` i stället för panik.

### Vem äger Lua-tillståndet?

Ett `LState` går inte att använda från flera goroutiner – precis som
världen. Därför tillhör det spelloopen: skript laddas och körs bara där
inne. Ett skript kan då anropa `mud.send` och läsa spelare utan lås.

## Kontraktet

- `Options.ScriptDir` – en mapp med `.lua`-filer (`""` = inga skript). De
  laddas när servern skapas, i namnordning, och laddas om (i spelloopen) när
  en fil läggs till, tas bort eller ändras – kontrolleras var `ReloadEvery`.
- En trasig fil hoppas över; de andra laddas ändå; inloggade administratörer
  får `MsgScriptLoadFailed` med felet, som nämner filen.
- Sandlådan: `os`, `io`, `dofile`, `loadfile`, `load`, `loadstring` och
  `require` är `nil`.
- API:t (en global tabell `mud`):
  - `mud.command(namn, function(player, args))` – ett nytt kommando;
  - `mud.behavior(namn, function(npc, room))` – ett beteende, används i
    NPC-filerna som `{"type": "script", "script": "namn"}` (`BehaviorSpec.Script`);
  - `mud.send(player, text)`, `mud.broadcast(room, text)`;
  - `mud.player(namn)` → `{name, room, hp, max_hp}` eller `nil`;
  - `mud.players(room)` → en lista med namn; `mud.roll(a, b)` → ett slumptal från a till b.
- Inbyggda kommandon går först: ett skript kan *lägga till* kommandon, inte ersätta dem.
- Varje anrop har en tidsgräns på `ScriptTimeout` (50 ms); ett skript som
  får fel eller kör för länge ger spelaren `MsgScriptError`, och servern fortsätter.
- `main.go`: `-scripts scripts`.

## Uppgift

`go get github.com/yuin/gopher-lua@v1.1.1` först. `scripts.go` är delvis
given (`newScripts` med API:t, `loadScripts`, `ScriptBehavior`,
`watchScripts`): skriv `run`, `LoadScripts` och `runScriptCommand`. Sedan
alternativet, fälten, `dispatch`, beteendetypen och flaggan.
`server/scripts/hello.lua` är ett exempel.

## Fundera på

- Varför får skript lägga till kommandon men inte ersätta `look`? När vore det användbart att ersätta?
- Vad skulle ett elakt skript fortfarande kunna göra? (Minne? `string.rep("x", 1e9)`?)
- API:t ger skripten namn (strängar), inte Go-pekare. Varför?
- Lua körs i spelloopen. Ett skript som tar 40 ms – vad gör det med tickfrekvensen? Vad skulle du göra åt det?

## Utmaningar

- `mud.on_tick(function() ... end)` – skript som gör något varje tick.
- `mud.give(player, item)` och `mud.move(player, room)` – vilka kontroller behöver de?
- Låt byggare (inte bara administratörer) ladda upp skript med ett `@script`-kommando.
- En minnesgräns: räkna storleken på tabellerna ett skript skapar.
