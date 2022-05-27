// Code generated by go-localize; DO NOT EDIT.
// This file was generated by robots at
// 2022-05-27 19:59:13.04395 +0200 CEST m=+0.006453101

package localizations

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

var localizations = map[string]string{
	"en.texts.achievement":                              ":trophy: Achievement :trophy: ",
	"en.texts.achievement_longest_audio_description":    "He managed to talk for **{{.seconds}} seconds ** straight. Respect!",
	"en.texts.achievement_longest_audio_title":          ":lungs: Apnea expert",
	"en.texts.achievement_most_audios_sent_description": "He sent **{{.audios}} audios**, please, give me a break!",
	"en.texts.achievement_most_audios_sent_title":       ":speaking_head: The spammer",
	"en.texts.achievement_random_description_1":         "You have been randomly selected to win this achievement. Congrats!",
	"en.texts.achievement_random_description_2":         "My creator didn't implement this correctly, so theorically I could end up giving this achievement to myself, ha ha ha",
	"en.texts.achievement_random_description_3":         "Even this achievement description is randomly chosen. Isn't that fascinating?",
	"en.texts.achievement_random_description_4":         "Because you are the best of the best",
	"en.texts.achievement_random_description_5":         "You won this achievement. Maybe next time other person wins it, ha ha ha",
	"en.texts.achievement_random_title":                 ":flushed: Because you deserve it!",
	"en.texts.download_link_title":                      "Download link",
	"en.texts.duration":                                 "Duration",
	"en.texts.hello":                                    ">>> :wave: Hey! I am **{{.botName}}**. Im capable of recording everything you say and transform it into an audio message.\n:flag_gb: I am configured to talk to you in english.\n:microphone2: To start recording, enter the channel **{{.voiceChannel}}**, wait until discord shows :green_circle: voice connected and start talking to me.\n:sound: When you get out the channel, I will send the audio message.\n\nMade with :heart: by Héctor <https://github.com/hectorgabucio/taterubot-dc> and tested by Aranchi and Raulio.",
	"en.texts.stats":                                    ">>> :chart_with_upwards_trend: **Monthly stats**: \n\n:earth_africa: Global stats:\n- {{.globalDuration}} seconds of audio sent\n- {{.globalAmount}} audio files recorded\n- Median duration of {{.globalMedianDuration}} seconds",
	"en.texts.stats-empty":                              ">>> :tired_face: Start sending voice messages to have stats!",
	"es.texts.achievement":                              ":trophy: Logro :trophy: ",
	"es.texts.achievement_longest_audio_description":    "Ha podido hablar durante **{{.seconds}} segundos ** de golpe. Respira un poco!",
	"es.texts.achievement_longest_audio_title":          ":lungs: Experto en aguantar la respiración",
	"es.texts.achievement_most_audios_sent_description": "Ha mandado **{{.audios}} audios**. Por favor, deja de darme trabajo!",
	"es.texts.achievement_most_audios_sent_title":       ":speaking_head: El metralletas",
	"es.texts.achievement_random_description_1":         "Has sido elegido aleatoriamente para ganar este premio. Felicidades!",
	"es.texts.achievement_random_description_2":         "Mi creador no me ha programado bien, así que teoricamente puedo acabar dándome este premio a mi mismo, :robot: ja ja ja :robot: ",
	"es.texts.achievement_random_description_3":         "Hasta esta descripción es aleatoria. No es eso fascinante?",
	"es.texts.achievement_random_description_4":         "Porque eres el mejor de los mejores, y punto.",
	"es.texts.achievement_random_description_5":         "Has ganado este premio, y ya. Puede que la próxima vez lo gane otra persona, ja ja ja :robot:",
	"es.texts.achievement_random_title":                 ":flushed: Porque te lo mereces, y porque me da la gana!",
	"es.texts.download_link_title":                      "Enlace de descarga",
	"es.texts.duration":                                 "Duración",
	"es.texts.hello":                                    ">>> :wave: Hola! Soy **{{.botName}}**. Soy un bot capaz de grabar lo que dices y transformarlo en mensajes de voz.\n:flag_es: Estoy configurado para responderte en castellano.\n:microphone2: Para empezar a grabar, entra en el canal **{{.voiceChannel}}**, espera a que Discord muestre :green_circle: *Voz conectada* y empieza a hablarme.\n:sound: Cuando salgas del mismo, mandaré el mensaje de voz.\n\nHecho con :heart: por Héctor <https://github.com/hectorgabucio/taterubot-dc> y probado por Aranchi y Raulio.",
	"es.texts.stats":                                    ">>> :chart_with_upwards_trend: **Estadísticas del mes**: \n\n:earth_africa: Estadísticas generales:\n- Un total de {{.globalDuration}} segundos enviados como audio\n- {{.globalAmount}} archivos de audio grabados\n- Duración media de {{.globalMedianDuration}} segundos",
	"es.texts.stats-empty":                              ">>> :tired_face: Empieza a mandar mensajes de voz para tener estadísticas!",
}

type Replacements map[string]interface{}

type Localizer struct {
	Locale         string
	FallbackLocale string
	Localizations  map[string]string
}

func New(locale string, fallbackLocale string) *Localizer {
	t := &Localizer{Locale: locale, FallbackLocale: fallbackLocale}
	t.Localizations = localizations
	return t
}

func (t Localizer) SetLocales(locale, fallback string) Localizer {
	t.Locale = locale
	t.FallbackLocale = fallback
	return t
}

func (t Localizer) SetLocale(locale string) Localizer {
	t.Locale = locale
	return t
}

func (t Localizer) SetFallbackLocale(fallback string) Localizer {
	t.FallbackLocale = fallback
	return t
}

func (t Localizer) GetWithLocale(locale, key string, replacements ...*Replacements) string {
	str, ok := t.Localizations[t.getLocalizationKey(locale, key)]
	if !ok {
		str, ok = t.Localizations[t.getLocalizationKey(t.FallbackLocale, key)]
		if !ok {
			return key
		}
	}

	// If the str doesn't have any substitutions, no need to
	// template.Execute.
	if strings.Index(str, "}}") == -1 {
		return str
	}

	return t.replace(str, replacements...)
}

func (t Localizer) Get(key string, replacements ...*Replacements) string {
	str := t.GetWithLocale(t.Locale, key, replacements...)
	return str
}

func (t Localizer) getLocalizationKey(locale string, key string) string {
	return fmt.Sprintf("%v.%v", locale, key)
}

func (t Localizer) replace(str string, replacements ...*Replacements) string {
	b := &bytes.Buffer{}
	tmpl, err := template.New("").Parse(str)
	if err != nil {
		return str
	}

	replacementsMerge := Replacements{}
	for _, replacement := range replacements {
		for k, v := range *replacement {
			replacementsMerge[k] = v
		}
	}

	err = template.Must(tmpl, err).Execute(b, replacementsMerge)
	if err != nil {
		return str
	}
	buff := b.String()
	return buff
}
