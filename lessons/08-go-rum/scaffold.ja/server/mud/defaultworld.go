package mud

// DefaultWorld is the village of Mudby. Feel free to add more rooms!
func DefaultWorld() *World {
	return NewWorld("torget",
		&Room{
			ID:          "torget",
			Name:        "マドビー広場",
			Description: "石畳の広場。中央に古い井戸があり、\n北の酒場から笑い声とジョッキの音が聞こえる。",
			Exits:       map[string]string{"north": "krogen", "east": "smedjan", "south": "porten", "west": "templet"},
		},
		&Room{
			ID:          "krogen",
			Name:        "酒場「陽気な竜亭」",
			Description: "エールと煙のにおいがする。階段が客室へ続いている。",
			Exits:       map[string]string{"south": "torget", "up": "gastrummet"},
		},
		&Room{
			ID:          "gastrummet",
			Name:        "小さな客室",
			Description: "狭いベッドと、広場に面した窓がひとつ。",
			Exits:       map[string]string{"down": "krogen"},
		},
		&Room{
			ID:          "smedjan",
			Name:        "鍛冶屋",
			Description: "炉の熱が押し寄せてくる。槌の音が壁に響いている。",
			Exits:       map[string]string{"west": "torget"},
		},
		&Room{
			ID:          "templet",
			Name:        "神殿",
			Description: "静かでひんやりとした神殿。祭壇の前でろうそくが揺れている。",
			Exits:       map[string]string{"east": "torget"},
		},
		&Room{
			ID:          "porten",
			Name:        "町の門",
			Description: "町の城壁にある重い木の門。南へ森に向かう道が続いている。",
			Exits:       map[string]string{"north": "torget"},
		},
	)
}
