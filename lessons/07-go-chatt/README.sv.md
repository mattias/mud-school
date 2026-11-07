# Lektion 07 – Chatt

Nu blir det socialt: `say hej` ska höras av alla andra. Det låter enkelt –
loopa över spelarna och skriv till deras anslutningar – men det döljer en
fälla som har fällt många riktiga MUD:ar.

## Fällan

Tänk dig att Ludvig har dålig uppkoppling (eller en buggig klient som slutat
läsa). Du skriver `say hej`. Din goroutine försöker skriva till Ludvigs
socket. Operativsystemets buffert för Ludvig är full, så `Write` **blockerar**
– tills Ludvig läser. Medan din goroutine väntar håller den kanske låset
`s.mu`. Nu kan ingen logga in, ingen kan skriva `who` ... *en* långsam spelare
har fryst hela servern.

Lösningen är ett klassiskt Go-mönster:

> **Varje anslutning har exakt en goroutine som skriver till den.** Alla andra
> lägger meddelanden i en **kanal** (spelarens utkorg) och går vidare direkt.

```
 Annas goroutine ──┐
 Bertils goroutine ┼──► [ Ludvigs utkorg (chan string) ] ──► Ludvigs skrivargoroutine ──► socket
 ...               ┘
```

## Mål

- Skapa och använda **kanaler**, både obuffrade och buffrade.
- Använda `select` med `default` för att skicka *utan att blockera*.
- Stänga en kanal och förstå vad `range` över en kanal gör.
- Förstå mottryck (backpressure): vad gör man när mottagaren inte hänger med?

## Go-koncept

```go
ch := make(chan string, 256)  // buffrad: rymmer 256 meddelanden innan send blockerar

ch <- "hej"                   // skicka (blockerar om bufferten är full)
msg := <-ch                   // ta emot (blockerar om kanalen är tom)

for msg := range ch { ... }   // ta emot tills kanalen stängs
close(ch)                     // bara SÄNDAREN stänger; skicka på stängd kanal = panik!

select {                      // välj det första som är redo ...
case ch <- msg:
    // skickat
default:
    // ... eller gör detta direkt om inget är redo
}
```

Kanalen `done chan struct{}` är ett idiom för "signal utan data":
`close(done)` väcker alla som väntar på `<-done`.

## Uppgift

1. **`player.go`**:
   - ge `Player` fälten `outbox chan string` (buffrad, t.ex. 256) och
     `done chan struct{}`,
   - starta en `writeLoop`-goroutine i `newPlayer` som är den **enda** som
     skriver till `conn`, och som stänger `done` när den är klar,
   - låt `Send` lägga meddelandet i utkorgen med `select`/`default`. Är
     utkorgen full: koppla ner spelaren (`p.conn.Close()`),
   - skriv `close()` som stänger utkorgen och väntar på `done`.
2. **`server.go`**:
   - skriv `broadcast(except *Player, text string)` som skickar till alla
     inloggade utom `except`,
   - meddela andra när någon ansluter (`MsgJoined`: `Bertil har anslutit.`) och lämnar
     (`MsgLeft`: `Bertil har lämnat.`) – oavsett om det sker med `quit` eller genom att
     koppla ner,
   - anropa `p.close()` innan anslutningen stängs, annars hinner `MsgGoodbye`
     kanske aldrig skickas.
3. **`commands.go`**: lägg till `say`. Avsändaren ser `MsgYouSay`
   (`Du säger: hej`), alla andra `MsgSays` (`Anna säger: hej`). Utan text:
   `MsgSayWhat` (`Säg vad?`).

Testet `TestSlowClientDoesNotBlockOthers` loggar in en spelare som aldrig
läser och låter en annan prata ~8 MB text. Med den naiva lösningen hänger
testet; med utkorgar går det igenom.

## Testa själv

Starta servern, två klienter (`./mudschool connect` eller din Rust-klient från lektion 06)
och prata med dig själv. Kör `./mudschool check --race` också.

## Fundera på

- Ordningen på `defer` i `handle` spelar roll. Vad händer om `conn.Close()`
  körs före `p.close()`?
- Vi kopplar ner en långsam spelare. Vilka andra strategier finns?
  (Släppa meddelanden? Slå ihop dem? Vad gör t.ex. ett chattprogram?)
- Ett alternativ till mutex + kanaler är att ha en *enda* goroutine som äger
  hela spelvärlden och som alla andra skickar kommandon till. Det kallas
  ibland "actor"-mönstret och vi använder det i fas 3 för spelloopen.

## Utmaningar

- `tell <namn> <text>` – viska till en enda spelare.
- `emote dansar` → `Anna dansar.`
