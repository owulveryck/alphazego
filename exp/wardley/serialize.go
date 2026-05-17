package wardley

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
)

// SerializeWTG2 retourne le texte WTG2 de la carte.
// Avec l'architecture WTG2-native, le texte est stocké tel quel.
func SerializeWTG2(s *State) string {
	return s.WTG2Text()
}

const playgroundBase = "https://owulveryck.github.io/wardleyToGo/?wtg2="

// PlaygroundURL retourne l'URL du playground wardleyToGo pour visualiser la carte.
// Le texte WTG2 est compressé en gzip puis encodé en base64url sans padding.
func PlaygroundURL(s *State) (string, error) {
	text := SerializeWTG2(s)

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(text)); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}

	encoded := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(buf.Bytes())
	return playgroundBase + encoded, nil
}
