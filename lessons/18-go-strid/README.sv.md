# Lektion 18 – Strid

Råttan i grottan visar tänderna. Vargen i Mörkskogen ylar. Dags att slåss.
Strid i en MUD är ett klassiskt jobb för spelloopen: när en strid väl har
börjat går den **en runda per tick** tills någon faller eller flyr.

```
> kill råtta
Du anfaller en grottråtta!
Du träffar en grottråtta (3 i skada).   ← tick
En grottråtta träffar dig (1 i skada).
Du träffar en grottråtta (2 i skada).   ← tick
Du har besegrat en grottråtta!
```

Det svåra är inte striden – det är att **testa** något som beror på tärningar
och tid. Lektionens tester fryser världen (`Tick: time.Hour`) och stegar den
framåt själva, ett tick i taget:

```go
srv.call(func() { srv.ticks++; srv.onTick() })   // exakt ett tick
```

Det är därför spelloopens tillstånd ligger på ett ställe: ett test kan styra
den som en debugger.

## Mål

- Modellera en strid som tillstånd i spelloopen (vem slåss mot vem) och föra den framåt varje tick.
- Skriva slump som är rättvis och testbar (`roll` med en injicerad `rand.Rand`).
- Skriva deterministiska tester för tidsberoende beteende.
- Utöka protokollet igen: `status.target`.

## Kontraktet

- `const PlayerMaxHP = 20`, `const DefaultRespawn = 50`,
  `var PlayerDamage = [2]int{1, 4}`; `Options.Respawn` (ticks, 0 = standard).
- `Player.HP`, `Player.MaxHP` (fullt vid inloggning) och `Player.target *NPC`.
- Varelsens fält från lektion 16: `HP` (max-hälsa), `Damage [2]int`,
  `Attackable`; under körning dess nuvarande hälsa, om den är besegrad och
  när den kommer tillbaka.
- Kommandon:
  - `kill <varelse>` (`k`, `attack`): `MsgAttackWhat`, `MsgNoSuchTarget`,
    `MsgCantAttack`, annars `MsgYouAttack` till dig och `MsgAttacks` till rummet.
  - `flee`: `MsgNotFighting`, `MsgFleeNowhere`, annars `MsgYouFlee` och en
    promenad genom en slumpad utgång. Striden är över.
- Varje tick, för varje spelare som slåss, i **namnordning**:
  1. spelaren slår för `roll(PlayerDamage)`: `MsgYouHit`;
  2. om varelsens hälsa är ≤ 0: `MsgYouKill` till dig, `MsgKills` till
     rummet, alla som slåss mot den slutar, den försvinner och kommer tillbaka
     efter `Respawn` ticks i sitt startrum med full hälsa (`MsgAppears`);
  3. annars slår varelsen tillbaka för `roll(n.Damage)`: `MsgHitsYou`;
  4. om din hälsa är ≤ 0: `MsgDies` till rummet, du vaknar i startrummet med
     full hälsa (`MsgYouDie` med rummets namn) och får en `look`.
- En varelse som någon slåss mot vandrar inte.
- Alla stridsmeddelanden har kind `combat`.
- `Status` får `HP`/`MaxHP` från spelaren och
  `Target *Target{Name, HP, MaxHP}` (json `target,omitempty`) under en strid.
  En strukturerad klient får en status efter varje runda.

## Uppgift

1. **`combat.go`** (ny): fyll i funktionerna.
2. **`loop.go`**: `onTick` hoppar över besegrade varelser och kör sedan
   `combatRound()` och `respawnCreatures()`.
3. **`npc.go`**: fälten för körning (nuvarande HP, besegrad, respawn-tick,
   startrum), som sätts i `AddNPCs`; `Wanderer` står still under en strid.
4. **`server.go`**/**`player.go`**: `Options.Respawn`, spelarens HP.
5. **`frame.go`**/**`rooms.go`**: `Target` i statusen.
6. Registrera `kill` och `flee`, och förkortningarna `k` och `attack`.

## Testa själv

```
./mudschool run server
```

Gå till grottan (s, s, s, s, s, s, e, s) och skriv `kill råtta`. Prova sedan
vargen i Mörkskogen – och `flee` i tid!

## Fundera på

- Varför stegar testerna världen själva istället för att sova och hoppas att
  tillräckligt många ticks har gått?
- Varför måste stridsrundan gå igenom spelarna i en fast ordning?
- Vad händer om två spelare slåss mot samma varelse? Vem får segern?

## Utmaningar

- Erfarenhetspoäng: besegrade varelser ger XP, och tillräckligt med XP höjer ditt max-HP.
- Vapen: att bära det rostiga svärdet (lektion 15) ger mer skada.
- Aggressiva varelser som anfaller spelare som kommer in.
