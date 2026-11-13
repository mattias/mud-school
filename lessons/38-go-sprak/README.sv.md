# Lektion 38 – Språk per spelare

Kursen talar tre språk, men din server talar ett: språket den byggdes på.
När den väl ligger på webben (lektion 42) kan en japansk och en svensk
spelare mötas i samma rum. I den här lektionen väljer varje spelare ett
språk, och världens texter följer med:

```
> language
Ditt språk: sv. Det finns: en, ja, sv. Välj med t.ex. language sv
> language en
Språk: en.
The Hall of Beginnings
A quiet hall with whitewashed walls. ...
[Introduktion 1/6] Welcome, Anna! This short tutorial teaches you the basics ...
```

Det som översätts är **världen**: introduktionen, och rum som har
översättningar. Serverns egna meddelanden (`messages.go`) stannar på
språket den byggdes på – att översätta dem är ett större jobb (se
utmaningarna).

## Mål

- Låt ett värde i JSON ha **två former**: en sträng, eller ett objekt med en
  sträng per språk – med en egen `UnmarshalJSON`.
- Fall tillbaka förnuftigt: spelarens språk, världens, den vanliga texten, vad som helst.
- Fyll i texter med `text/template`: `Välkommen, {{.Name}}!`.
- Låt det gamla filformatet fungera: det som var en sträng är det fortfarande.

## Go-koncept

### `json.Unmarshaler`

```go
type Text map[string]string

func (t *Text) UnmarshalJSON(data []byte) error {
	var plain string
	if err := json.Unmarshal(data, &plain); err == nil {
		*t = Text{"": plain}
		return nil
	}
	...
}
```

`encoding/json` kollar om en typ har en `UnmarshalJSON`-metod – ett
gränssnitt, `json.Unmarshaler` – och låter den avkoda sig själv. En vanlig
sträng sparas under nyckeln `""`, så `"text": "Skriv look."` fungerar
fortfarande, och `"text": {"en": "Type look.", "sv": "Skriv look."}` fungerar
också. `MarshalJSON` gör det omvända, så när världen sparas (lektion 36)
skrivs en vanlig sträng tillbaka som en sträng.

### En metod på en map-typ

`Text` är en map med metoder. `t.In("sv", "en")` letar efter `sv`, sedan
världens språk, sedan `""`, och till sist det första språket i bokstavsordning
– en text på något språk är bättre än ingen.

### `text/template`

```go
tmpl, err := template.New("text").Option("missingkey=error").Parse(s)
...
tmpl.Execute(&b, TextData{Name: "Anna", Step: 2, Steps: 6})
```

Mallarna kontrolleras när världen laddas – att köra dem en gång med en tom
`TextData` hittar `{{.Nmae}}` och trasiga klamrar – så ett stavfel blir ett
fel i filen, inte en överraskning för en spelare.

## Kontraktet

- `Text` (`map[string]string`) med `UnmarshalJSON` (en sträng → `{"": s}`; ett
  objekt; annars ett fel med `MsgErrText`), `MarshalJSON`, `In(lang, fallback)`,
  `Render(lang, fallback, TextData)` och `check()` (`MsgErrTemplate` med språket).
  `TextData{Name, Step, Steps}`.
- `TutorialStep.Text` och `Tutorial.Done` är `Text`. En text som inte är en
  giltig mall är ett laddningsfel: `MsgErrTutorialText` med stegets nummer
  (`len(Steps)+1` för `done`). Ett steg vars text är tom på alla språk
  (`""`, `{"en": ""}`, `null`) har ingen text: `MsgErrTutorialStep`, som i lektion 37.
- En världsfil får ha `"language"` (`World.Language`: språket på dess vanliga
  texter), och rum får ha `"translations": {"sv": {"name": ..., "description": ...}}`
  (`Room.Translations`, `RoomText`).
- `(w *World) Languages()` – alla språk världen har texter på, sorterade.
- `language` visar `MsgLanguage` (spelarens språk, eller världens, och
  listan); `language xx` väljer det (`MsgLanguageSet`, sedan en look och
  introduktionens steg igen), eller `MsgLanguageUnknown`.
- Rumsbeskrivningen och status-framens rumsnamn använder spelarens språk;
  introduktionens texter renderas med spelarens namn och steg.
- `SavedPlayer.Language`, sparad genom migrering 5.
- `DefaultWorld` använder `tutorial.all.json` (given: introduktionen på sv, en
  och ja) och sätter `Language` till `CourseLanguage` (i `messages.go`).
- `worldSnapshot` behåller `Language`; översättningar och `Text` sparas som de lästes.

## Uppgift

1. `i18n.go` är delvis given (typerna, `empty`, `sortedKeys`): skriv de andra `Text`-metoderna,
   `Languages`, `roomText`, `render` och `cmdLanguage`.
2. `tutorial.go`: `Text` i stället för `string`, mallkontrollen, och
   `s.render` där texterna visas.
3. `world.go`: `Room.Translations`, `World.Language`, `worldFile.Language`.
   `rooms.go`: `roomText` i `describeRoom` och `sendStatus`.
4. `player.go`: `lang string`. `store.go`: migrering 5 och `SavedPlayer.Language`.
   `account.go`: spara och återställ den.
5. `defaultworld.go`: `tutorial.all.json` och `w.Language = CourseLanguage`.
   `commands.go`: `language`. `olc.go`: `Language` i `worldSnapshot`.
   `copyover.go`: `Lang` i `CopyoverPlayer` (JSON `lang`), `PrepareCopyover` och `adopt`.

## Fundera på

- `@desc` (lektion 36) ändrar ett rums egen beskrivning. Vad händer med dess
  översättningar? Vad borde hända?
- Varför är ett mallfält som saknas ett fel när världen laddas, medan en mall
  som misslyckas när spelet körs visas som vanlig text?
- Webbklienten (lektion 39) skulle kunna skicka `language` själv, från
  webbläsarens språk. Var skulle det passa in?
- Vilket är lättast att översätta: en mening ihopsatt med `%s`, eller en mall?

## Utmaningar

- Översätt serverns meddelanden: en katalog per språk (de tre `messages.go`
  är en bra början) och `p.Msg("MsgWelcomeName", name)`.
- Saker och varelser med översatta namn.
- `@translate sv <text>`: byggare översätter rummet de står i.
