package i18n

import (
	"io"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

type Entries map[string]interface{}

// internal error  that has key of the error string, not actual string
// to allow for localization
type Error struct {
	key string
	obj any
}

func (e *Error) Error() string {
	return e.key
}

func (e *Error) GetKey() string {
	return e.key
}

func (e *Error) GetData() any {
	return e.obj
}
func NewError(key string, obj any) *Error {
	return &Error{key: key, obj: obj}
}

type Localizer struct {
	fallback         language.Tag
	fallbackEntries  Entries
	current_lang     language.Tag
	Printer          *message.Printer
	builder          *catalog.Builder
	TemplatesDefined []string
}

func NewLocalizer(lang language.Tag, fallback language.Tag) *Localizer {

	localizer := &Localizer{
		current_lang: lang,
		fallback:     fallback,
	}

	localizer.builder = catalog.NewBuilder(catalog.Fallback(fallback))
	localizer.Printer = message.NewPrinter(lang, message.Catalog(localizer.builder))
	localizer.TemplatesDefined = make([]string, 0)

	return localizer
}

func (l *Localizer) SetLanguage(tag language.Tag) {
	l.current_lang = tag
	l.Printer = message.NewPrinter(tag, message.Catalog(l.builder))

}

func (l *Localizer) GetLanguage() language.Tag {
	return l.current_lang
}

func (l *Localizer) addMessage(tag language.Tag, key string, msg interface{}) {
	switch typed_msg := msg.(type) {
	case string:
		// check if this is define of a sub-template
		s := strings.TrimLeft(typed_msg, " \t\n{-")
		if strings.HasPrefix(s, "define") {
			// this is a define of sub-template - remember the key
			l.TemplatesDefined = append(l.TemplatesDefined, key)
		}
		l.builder.SetString(tag, key, typed_msg)
	case catalog.Message:
		l.builder.Set(tag, key, typed_msg)
	case []catalog.Message:
		l.builder.Set(tag, key, typed_msg...)
	}
}

func (l *Localizer) loadEntries(entries map[language.Tag]Entries) {

	for tag, msgs := range entries {
		for key, original_mgs := range l.fallbackEntries {
			var msg interface{}
			if translated, ok := msgs[key]; ok {
				msg = translated
				if tag != l.fallback {
					delete(msgs, key)
				}
			} else {
				//use original for this message
				msg = original_mgs
			}
			l.addMessage(tag, key, msg)
		}
		// if we have extra strings user wants to handle - add them too
		if tag != l.fallback {
			for key, translated := range msgs {
				l.addMessage(tag, key, translated)
			}
		}
	}
}

func (l *Localizer) AddUpdateTranslation(lang language.Tag, entries Entries) {

	if lang == l.fallback && l.fallbackEntries == nil {
		l.fallbackEntries = entries
	} else if lang == l.fallback {
		// this is update to fallback entries
		for key, msg := range entries {
			l.fallbackEntries[key] = msg
		}
	}

	dict := map[language.Tag]Entries{
		lang: entries,
	}
	l.loadEntries(dict)

}

// Sprintf is like fmt.Sprintf, but using language-specific formatting.
func (l *Localizer) Sprintf(key string, a ...interface{}) string {
	return l.Printer.Sprintf(key, a...)
}

// Fprintf is like fmt.Fprintf, but using language-specific formatting.
func (l *Localizer) Fprintf(w io.Writer, key string, a ...interface{}) (n int, err error) {
	return l.Printer.Fprintf(w, key, a...)
}

// Printf is like fmt.Printf, but using language-specific formatting.
func (l *Localizer) Printf(key string, a ...interface{}) (n int, err error) {
	return l.Printer.Printf(key, a...)
}
