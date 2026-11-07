# Lektion 22 – Testa på riktigt

`ParseCommand` har följt dig sedan lektion 4. Den har tester. Den fungerar. Eller…?

Skriv `look`, en **tabb** och `sword` i `./mudschool connect` (eller telnet). Servern svarar att den inte
känner till kommandot `look	sword` – tabben blev en del av verbet. Inget av
testerna du har sett fångade det, för ingen kom på att skriva just det exemplet.

Det är gränsen för exempelbaserade tester: de testar bara det du kommer på.
I den här lektionen låter du datorn komma på exemplen i stället.

- **Fuzzing**: Go genererar miljontals indata, styrt av vilka kodvägar de
  når, och sparar varje indata som bryter mot en regel.
- **Egenskapsbaserade tester**: i stället för "den här indatan ger den här
  utdatan" skriver du regler som gäller för *all* indata.

## Mål

- Skriva ett fuzz-test: `func FuzzXxx(f *testing.F)`, frön med `f.Add`, testet med `f.Fuzz`.
- Köra fuzzern: `go test -fuzz=...` och läsa vad den hittade.
- Tänka i **invarianter** (egenskaper): vad måste alltid vara sant?
- Förstå `testdata/fuzz/`: indata som en gång fallerade blir regressionstester för alltid.
- Laga buggen med `unicode.IsSpace` – blanktecken är mer än `' '`.

## Go-koncept

### Ett fuzz-test

```go
func FuzzParseCommand(f *testing.F) {
	f.Add("look")                 // frön: vanliga exempel att börja från
	f.Add("Get Sword from chest")
	f.Fuzz(func(t *testing.T, line string) {
		c := ParseCommand(line)
		if strings.IndexFunc(c.Verb, unicode.IsSpace) >= 0 {
			t.Errorf("the verb %q contains whitespace", c.Verb)
		}
		// ... fler regler ...
	})
}
```

Det ligger i en vanlig `_test.go`-fil. Ett vanligt `go test` kör bara fröna –
så det är också ett vanligt test. Med `-fuzz` börjar Go mutera:

```
cd workspace/server
go test -fuzz=FuzzParseCommand -fuzztime=30s ./mud/
```

När den hittar något **krymper** den indatan till den minsta som fortfarande
fallerar och sparar den i `mud/testdata/fuzz/FuzzParseCommand/`. Committa den
mappen: från och med nu kör varje `go test` den indatan också.

### Vilka regler?

En bra invariant är något du är *säker* på utan att känna till indatan:

- verbet innehåller aldrig blanktecken;
- `Rest` har inga blanktecken i ändarna;
- `Args` är exakt orden i `Rest` (`strings.Fields(Rest)`);
- verbet är tomt om och endast om raden är tom;
- verbet är gemener (för giltig UTF-8).

Lägg märke till vad som *inte* står på listan: "verbet är ett känt kommando".
Tolken känner inte till kommandona – det är serverns sak.

### Blanktecken

```go
unicode.IsSpace(' ')      // true
unicode.IsSpace('\t')     // true
unicode.IsSpace(' ') // true: ett hårt mellanslag
unicode.IsSpace('　') // true: det ideografiska mellanslag ett japanskt tangentbord skriver
```

`strings.Fields` och `strings.TrimSpace` använder redan `unicode.IsSpace`.
Det gör inte `strings.Cut(line, " ")`. `strings.IndexFunc(line, unicode.IsSpace)`
hittar det första blanktecknet av alla sorter.

## Kontraktet

- Din egen `server/mud/command_fuzz_test.go` med `FuzzParseCommand` – frön
  (`f.Add`) och `f.Fuzz` med reglerna ovan. (Inte i en `mudschool_`-fil: de
  tillhör kursen och byts ut.)
- `ParseCommand` delar av verbet vid det första blanktecknet av *alla* sorter.

Kursens test kontrollerar att ditt fuzz-test finns, provar några kluriga
rader och kör ett eget litet egenskapsbaserat test med 5000 slumpade rader.

## Uppgift

1. Skriv `FuzzParseCommand` med några vanliga frön (*inte* tabben – låt
   fuzzern hitta buggen själv).
2. Kör fuzzern. Hur lång tid tog det? Hur såg den krympta indatan ut?
3. Titta i `testdata/fuzz/FuzzParseCommand/`.
4. Laga `ParseCommand`. Kör fuzzern igen i en minut.

## Fundera på

- Ditt fuzz-test hittade på sekunder en bugg som testerna från lektion 4
  aldrig hittade. Betyder det att exempeltesterna var meningslösa?
- Vad mer i servern tar emot indata från främlingar? (`ValidateName`,
  världsinläsningen, JSON från en grafisk klient …) Vilka regler skulle du
  fuzza dem med?
- Varför är fuzzing särskilt viktigt för ett program på internet?
- Fuzzing hittar krascher och brutna regler – men kan den säga att `look`
  visar *rätt* rum?

## Utmaningar

- Fuzza `ValidateName` och `NormalizeName`: gäller `NormalizeName(NormalizeName(x)) == NormalizeName(x)`?
- Fuzza `LoadWorld` med slumpad JSON: den får returnera ett fel, men aldrig få panik.
- I Rust: lägg till `proptest` som dev-dependency och skriv ett egenskapstest
  för `parse_line` – vilken sträng som helst blir *någon* `ServerMsg`, utan panik.
- En bot: ett litet Go-program som loggar in och går omkring på måfå. Kör tio
  av dem mot din server med `-race` i en minut.
