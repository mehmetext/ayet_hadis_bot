package output

import (
	"fmt"
	"io"

	"github.com/example/ayet-hadis-bot/internal/content"
)

func PrintContent(writer io.Writer, verse content.Verse, hadith content.Hadith) error {
	_, err := fmt.Fprintf(writer, "Ayet [%s]\n%s\n\nHadis [%s]\n%s\n", verse.Reference, verse.Text, hadith.Reference, hadith.Text)
	return err
}
