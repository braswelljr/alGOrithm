package core

import (
	"fmt"
	"math/big"
	"math/bits"
	"slices"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/security/ciphers"
	"github.com/braswelljr/alGOrithm/pkg/security/cryptography"
	"github.com/braswelljr/alGOrithm/pkg/security/hashing"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// ciphersEntries is every entry of the ciphers category.
func ciphersEntries() []Entry {
	return concat(
		[]Entry{
			caesarCipherEntry(),
			hillCipherEntry(),
			rot13Entry(),
			vigenereCipherEntry(),
		},
		classicalEntries(),
	)
}

// cryptographyEntries is every entry of the cryptography category.
func cryptographyEntries() []Entry {
	return modernEntries()
}

// hashingEntries is every entry of the hashing category.
func hashingEntries() []Entry {
	return hashFunctionEntries()
}

func ciphersText(in Input) error {
	if in.Text == "" {
		return inputError("give the text to encrypt")
	}
	return nil
}

func caesarCipherEntry() Entry {
	caesar := entry("caesar-cipher", "Caesar Cipher", "security/ciphers/caesar.go", "Caesar")
	caesar.Summary, caesar.Time, caesar.Space, caesar.Input, caesar.Params = "Shift every letter forward by a fixed number of places (A becomes D for 3).", "O(n)", "O(n)", Text, []string{"shift"}
	caesar.Example = &Input{Text: "Attack at dawn", Shift: 3}
	caesar.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := ciphersText(in); err != nil {
			return Output{}, err
		}
		return Output{Text: ciphers.Caesar(in.Text, in.Shift)}, nil
	}
	return caesar
}

func rot13Entry() Entry {
	rot := entry("rot13", "ROT13", "security/ciphers/caesar.go", "ROT13")
	rot.Summary, rot.Time, rot.Space, rot.Input, rot.Example = "Shift every letter by 13, so applying it twice gives the text back.", "O(n)", "O(n)", Text, &Input{Text: "Why did the chicken cross the road?"}
	rot.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := ciphersText(in); err != nil {
			return Output{}, err
		}
		return Output{Text: ciphers.ROT13(in.Text)}, nil
	}
	return rot
}

func vigenereCipherEntry() Entry {
	vig := entry("vigenere-cipher", "Vigenère Cipher", "security/ciphers/vigenere.go", "Vigenere")
	vig.Summary, vig.Time, vig.Space, vig.Input, vig.Params = "Shift each letter by the letter of a repeating keyword under it.", "O(n)", "O(n)", Text, []string{"key"}
	vig.Example = &Input{Text: "ATTACKATDAWN", Key: "LEMON"}
	vig.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := ciphersText(in); err != nil {
			return Output{}, err
		}
		if strings.TrimSpace(in.Key) == "" {
			return Output{}, inputError("give a keyword")
		}
		return Output{Text: ciphers.Vigenere(in.Text, in.Key)}, nil
	}
	return vig
}

func hillCipherEntry() Entry {
	hill := entry("hill-cipher", "Hill Cipher", "security/ciphers/hill.go", "Hill")
	hill.Summary, hill.Time, hill.Space, hill.Input = "Multiply blocks of letters by a key matrix modulo 26; decrypt with its inverse.", "O(n * k)", "O(n)", Text
	hill.Options = []Option{{"size", "key size", 2, 1, 5}, {"seed", "key seed", 1, 0, 1_000_000}}
	hill.Example = &Input{Text: "Help me, Hill", Options: map[string]int{"size": 2, "seed": 1}}
	hill.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := ciphersText(in); err != nil {
			return Output{}, err
		}
		k := in.opt("size", 2)
		if k < 1 || k > 5 {
			return Output{}, inputError("the key size is 1 to 5")
		}
		key := ciphers.HillKey(k, int64(in.opt("seed", 1)))
		enc, _ := ciphers.Hill(in.Text, key)
		dec, _ := ciphers.HillDecrypt(enc, key)
		cells := flatten(key)
		return Output{Text: enc, Lines: []string{"key " + fmt.Sprint(key), "decrypts to " + dec}, Table: &TableOutput{Rows: k, Cols: k, Cells: cells}}, nil
	}
	return hill
}

// crackSample is plain English long enough for the Vigenère attacks.
const crackSample = `It was the best of times, it was the worst of times, it was the age of wisdom, it was the age of foolishness,
it was the epoch of belief, it was the epoch of incredulity, it was the season of Light, it was the season of Darkness,
it was the spring of hope, it was the winter of despair, we had everything before us, we had nothing before us,
we were all going direct to Heaven, we were all going direct the other way.`

func needText(in Input) error {
	if in.Text == "" {
		return inputError("give some text")
	}
	return nil
}

func needKey(in Input, what string) error {
	if strings.TrimSpace(in.Key) == "" {
		return inputError("give " + what)
	}
	return nil
}

// keyedCipher is an entry for a cipher taking text and a keyword, traced.
func keyedCipher(id, name, file, fn, summary, what string, ex Input, f func(text, key string, r trace.Recorder) (string, error)) Entry {
	e := entry(id, name, file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Params, e.Traced, e.Example = summary, "O(n)", "O(n)", Text, []string{"key"}, true, &ex
	e.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := needText(in); err != nil {
			return Output{}, err
		}
		if err := needKey(in, what); err != nil {
			return Output{}, err
		}
		out, err := f(in.Text, in.Key, r)
		return Output{Text: out}, wrap(err)
	}
	e.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("ciphers.%s(%q, %q)", fn, in.Text, in.Key)
	}
	return e
}

func classicalEntries() []Entry {
	atbash := entry("atbash", "Atbash", "security/ciphers/atbash.go", "Atbash")
	atbash.Summary, atbash.Time, atbash.Space, atbash.Input, atbash.Traced = "Mirror the alphabet: A becomes Z, B becomes Y. Its own inverse.", "O(n)", "O(n)", Text, true
	atbash.Example = &Input{Text: "Hello, World"}
	atbash.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := needText(in); err != nil {
			return Output{}, err
		}
		return Output{Text: ciphers.TraceAtbash(in.Text, r)}, nil
	}
	atbash.Call = func(in Input) (string, string) { return "", fmt.Sprintf("ciphers.Atbash(%q)", in.Text) }

	affine := entry("affine-cipher", "Affine Cipher", "security/ciphers/affine.go", "Affine")
	affine.Summary, affine.Time, affine.Space, affine.Input, affine.Traced = "Map each letter x to (a*x + b) mod 26; a must share no factor with 26.", "O(n)", "O(n)", Text, true
	affine.Options = []Option{{"a", "multiplier a", 5, 1, 25}, {"b", "shift b", 8, 0, 25}}
	affine.Example = &Input{Text: "AFFINE CIPHER", Options: map[string]int{"a": 5, "b": 8}}
	affine.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := needText(in); err != nil {
			return Output{}, err
		}
		out, err := ciphers.TraceAffine(in.Text, in.opt("a", 5), in.opt("b", 8), r)
		return Output{Text: out}, wrap(err)
	}
	affine.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("ciphers.Affine(%q, %d, %d)", in.Text, in.opt("a", 5), in.opt("b", 8))
	}

	rail := entry("rail-fence", "Rail Fence Cipher", "security/ciphers/railfence.go", "RailFence")
	rail.Summary, rail.Time, rail.Space, rail.Input, rail.Traced = "Write the text in a zigzag over a number of rails, then read the rails one after another.", "O(n)", "O(n)", Text, true
	rail.Options = []Option{{"rails", "rails", 3, 2, 20}}
	rail.Example = &Input{Text: "WEAREDISCOVEREDFLEEATONCE", Options: map[string]int{"rails": 3}}
	rail.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := needText(in); err != nil {
			return Output{}, err
		}
		return Output{Text: ciphers.TraceRailFence(in.Text, in.opt("rails", 3), r)}, nil
	}
	rail.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("ciphers.RailFence(%q, %d)", in.Text, in.opt("rails", 3))
	}

	columnar := entry("columnar-transposition", "Columnar Transposition", "security/ciphers/columnar.go", "Columnar")
	columnar.Summary, columnar.Time, columnar.Space, columnar.Input, columnar.Params = "Write the text in rows under a keyword, then read the columns in the keyword's alphabetical order.", "O(n log k)", "O(n)", Text, []string{"key"}
	columnar.Example = &Input{Text: "WEAREDISCOVEREDFLEEATONCE", Key: "ZEBRAS"}
	columnar.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := needText(in); err != nil {
			return Output{}, err
		}
		if err := needKey(in, "a keyword"); err != nil {
			return Output{}, err
		}
		out, err := ciphers.Columnar(in.Text, in.Key)
		return Output{Text: out}, wrap(err)
	}
	columnar.Call = func(in Input) (string, string) { return "", fmt.Sprintf("ciphers.Columnar(%q, %q)", in.Text, in.Key) }

	playfair := entry("playfair", "Playfair Cipher", "security/ciphers/playfair.go", "Playfair")
	playfair.Summary, playfair.Time, playfair.Space, playfair.Input, playfair.Params = "Encrypt pairs of letters by their places in a 5x5 square built from a keyword (J counts as I).", "O(n)", "O(n)", Text, []string{"key"}
	playfair.Example = &Input{Text: "hide the gold in the tree stump", Key: "playfair example"}
	playfair.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := needText(in); err != nil {
			return Output{}, err
		}
		out, sq := ciphers.Playfair(in.Text, in.Key)
		lines := make([]string, 0, 5)
		for _, row := range sq {
			lines = append(lines, strings.Join(strings.Split(string(row[:]), ""), " "))
		}
		return Output{Text: out, Lines: lines}, nil
	}
	playfair.Call = func(in Input) (string, string) { return "", fmt.Sprintf("ciphers.Playfair(%q, %q)", in.Text, in.Key) }

	otp := entry("one-time-pad", "One-Time Pad", "security/ciphers/onetimepad.go", "OneTimePad")
	otp.Summary, otp.Time, otp.Space, otp.Input = "XOR the message with a truly random pad as long as it, used once: unbreakable, and impractical.", "O(n)", "O(n)", TwoTexts
	otp.Example = &Input{Text: "attack at dawn", Other: "qX7#vLp0z!Kd9wR2"}
	otp.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := needText(in); err != nil {
			return Output{}, err
		}
		out, err := ciphers.OneTimePad([]byte(in.Text), []byte(in.Other))
		return Output{Text: out}, wrap(err)
	}
	otp.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("ciphers.OneTimePad([]byte(%q), []byte(%q))", in.Text, in.Other)
	}

	entries := []Entry{
		atbash, affine, rail, columnar, playfair, otp,
		keyedCipher("autokey-cipher", "Autokey Cipher", "security/ciphers/autokey.go", "Autokey",
			"Vigenère with a key that runs on into the plaintext itself, so it never repeats.", "a keyword",
			Input{Text: "attackatdawn", Key: "QUEENLY"}, ciphers.TraceAutokey),
		keyedCipher("beaufort-cipher", "Beaufort Cipher", "security/ciphers/beaufort.go", "Beaufort",
			"Subtract each letter from the key letter (key - plain): the same operation encrypts and decrypts.", "a keyword",
			Input{Text: "DEFENDTHEEASTWALLOFTHECASTLE", Key: "FORTIFICATION"}, ciphers.TraceBeaufort),
		keyedCipher("substitution-cipher", "Substitution Cipher", "security/ciphers/substitution.go", "Substitution",
			"Replace each letter by the one in its place in a scrambled alphabet: 26! keys, broken by letter frequencies.", "a 26-letter alphabet",
			Input{Text: "flee at once", Key: "ZEBRASCDFGHIJKLMNOPQTUVWXY"}, ciphers.TraceSubstitution),
		enigmaEntry(),
	}
	return append(entries, attackEntries()...)
}

// parseEnigma reads "I II III ABC DEF AB CD": the rotors, positions, rings and
// plugboard pairs.
func parseEnigma(key string) (ciphers.EnigmaConfig, error) {
	f := strings.Fields(key)
	if len(f) < 3 {
		return ciphers.EnigmaConfig{}, inputError(`give the setup as "I II III AAA AAA AB CD": three rotors, then positions, rings and plugboard pairs`)
	}
	c := ciphers.EnigmaConfig{Rotors: [3]string{f[0], f[1], f[2]}}
	if len(f) > 3 {
		c.Positions = f[3]
	}
	if len(f) > 4 {
		c.Rings = f[4]
	}
	if len(f) > 5 {
		c.Plugboard = strings.Join(f[5:], " ")
	}
	return c, nil
}

func enigmaEntry() Entry {
	e := entry("enigma", "Enigma Machine", "security/ciphers/enigma.go", "Enigma")
	e.Summary, e.Time, e.Space, e.Input, e.Params = "The German Enigma I: three stepping rotors, a reflector and a plugboard; the setup is \"rotors positions rings pairs\".", "O(n)", "O(n)", Text, []string{"key"}
	e.Example = &Input{Text: "AAAAA", Key: "I II III AAA AAA"}
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := needText(in); err != nil {
			return Output{}, err
		}
		c, err := parseEnigma(in.Key)
		if err != nil {
			return Output{}, err
		}
		out, err := ciphers.Enigma(in.Text, c)
		return Output{Text: out}, wrap(err)
	}
	e.Call = func(in Input) (string, string) {
		c, err := parseEnigma(in.Key)
		if err != nil {
			return "", ""
		}
		return "", fmt.Sprintf("ciphers.Enigma(%q, ciphers.EnigmaConfig{Rotors: [3]string{%q, %q, %q}, Positions: %q, Rings: %q, Plugboard: %q})",
			in.Text, c.Rotors[0], c.Rotors[1], c.Rotors[2], c.Positions, c.Rings, c.Plugboard)
	}
	return e
}

func alphabetLabels() []string {
	l := make([]string, 26)
	for i := range l {
		l[i] = string(rune('A' + i))
	}
	return l
}

var ciphersMaxLen = Option{"max", "longest key to try", 12, 1, 40}

func caesarBruteForceEntry() Entry {
	brute := entry("caesar-brute-force", "Caesar Brute Force", "security/ciphers/attacks.go", "CaesarBruteForce")
	brute.Summary, brute.Time, brute.Space, brute.Input = "Try all 26 Caesar shifts and rank them by how English the letters look (chi-squared).", "O(26n)", "O(26n)", Text
	brute.Example = &Input{Text: "Wkh txlfn eurzq ira mxpsv ryhu wkh odcb grj"}
	brute.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := needText(in); err != nil {
			return Output{}, err
		}
		g := ciphers.CaesarBruteForce(in.Text)
		lines := make([]string, len(g))
		for i, x := range g {
			lines[i] = x.String()
		}
		return Output{Text: g[0].Text, Number: ptr(g[0].Shift), Lines: lines}, nil
	}
	brute.Call = func(in Input) (string, string) { return "", fmt.Sprintf("ciphers.CaesarBruteForce(%q)[0]", in.Text) }
	return brute
}

func frequencyAnalysisEntry() Entry {
	freq := entry("frequency-analysis", "Frequency Analysis", "security/ciphers/attacks.go", "FrequencyAnalysis")
	freq.Summary, freq.Time, freq.Space, freq.Input = "Count the letters and pick the Caesar shift whose letter frequencies best match English.", "O(26n)", "O(1)", Text
	freq.Example = &Input{Text: ciphers.Caesar(crackSample, 7)}
	freq.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := needText(in); err != nil {
			return Output{}, err
		}
		counts, shift := ciphers.FrequencyAnalysis(in.Text)
		return Output{
			Text:   ciphers.Caesar(in.Text, -shift),
			Number: ptr(shift),
			Table:  &TableOutput{Rows: 1, Cols: 26, Cells: counts[:], RowLabels: []string{"count"}, ColLabels: alphabetLabels()},
		}, nil
	}
	freq.Call = func(in Input) (string, string) { return "", fmt.Sprintf("ciphers.FrequencyAnalysis(%q)", in.Text) }
	return freq
}

func vigenereKeyLengthEntry() Entry {
	keyLen := entry("vigenere-key-length", "Vigenère Key Length", "security/ciphers/attacks.go", "VigenereKeyLength")
	keyLen.Summary, keyLen.Time, keyLen.Space, keyLen.Input = "Find a Vigenère key's length from repeated trigrams (Kasiski) and the columns' index of coincidence (Friedman).", "O(n * max + r^2)", "O(n)", Text
	keyLen.Options = []Option{ciphersMaxLen}
	keyLen.Example = &Input{Text: ciphers.Vigenere(crackSample, "DICKENS"), Options: map[string]int{"max": 12}}
	keyLen.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := needText(in); err != nil {
			return Output{}, err
		}
		ks := ciphers.VigenereKeyLength(in.Text, in.opt("max", 12))
		t := &TableOutput{Cols: 2, ColLabels: []string{"Kasiski repeats", "IC x 1000"}}
		for _, k := range ks {
			t.Rows++
			t.Cells = append(t.Cells, k.Kasiski, int(k.IC*1000+0.5))
			t.RowLabels = append(t.RowLabels, "length "+strconv.Itoa(k.Length))
		}
		return Output{Number: ptr(ks[0].Length), Table: t}, nil
	}
	keyLen.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("ciphers.VigenereKeyLength(%q, %d)[0]", in.Text, in.opt("max", 12))
	}
	return keyLen
}

func vigenereCrackEntry() Entry {
	crack := entry("vigenere-crack", "Vigenère Crack", "security/ciphers/attacks.go", "VigenereCrack")
	crack.Summary, crack.Time, crack.Space, crack.Input = "Break a Vigenère cipher: find the key length, then each key letter as the best Caesar shift of its column.", "O(n * max)", "O(n)", Text
	crack.Options = []Option{ciphersMaxLen}
	crack.Example = &Input{Text: ciphers.Vigenere(crackSample, "DICKENS"), Options: map[string]int{"max": 12}}
	crack.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := needText(in); err != nil {
			return Output{}, err
		}
		key, plain := ciphers.VigenereCrack(in.Text, in.opt("max", 12))
		return Output{Text: plain, Lines: []string{"key: " + key}}, nil
	}
	crack.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("ciphers.VigenereCrack(%q, %d)", in.Text, in.opt("max", 12))
	}
	return crack
}

func attackEntries() []Entry {
	return []Entry{
		caesarBruteForceEntry(),
		frequencyAnalysisEntry(),
		vigenereKeyLengthEntry(),
		vigenereCrackEntry(),
	}
}

// hexOf decodes s as exactly n bytes of hex.
func hexOf(s string, n int, what string) ([]byte, error) {
	b, err := cryptography.HexDecode(strings.TrimSpace(s))
	if err != nil || len(b) != n {
		return nil, inputError(fmt.Sprintf("give %s as %d hex digits", what, 2*n))
	}
	return b, nil
}

// hexCall writes a call whose array result is printed in hex.
func hexCall(expr string) (string, string) {
	return "sum := " + expr + "\n\t", "cryptography.Hex(sum[:])"
}

func modernEntries() []Entry {
	sha := entry("sha-256", "SHA-256", "security/cryptography/sha256.go", "SHA256")
	sha.Summary, sha.Time, sha.Space, sha.Input = "Hash text to 256 bits in 64 rounds of mixing eight words; shows a..h after every round of the first block.", "O(n)", "O(n)", Text
	sha.Example = &Input{Text: "abc"}
	sha.Run = func(in Input, _ trace.Recorder) (Output, error) {
		sum := cryptography.SHA256([]byte(in.Text))
		rounds := cryptography.SHA256Rounds([]byte(in.Text))
		lines := make([]string, len(rounds))
		for i, w := range rounds {
			lines[i] = fmt.Sprintf("round %2d: %08x", i, w[:])
		}
		return Output{Text: cryptography.Hex(sum[:]), Lines: lines}, nil
	}
	sha.Call = func(in Input) (string, string) {
		return hexCall(fmt.Sprintf("cryptography.SHA256([]byte(%q))", in.Text))
	}

	md5 := entry("md5", "MD5", "security/cryptography/md5.go", "MD5")
	md5.Summary, md5.Time, md5.Space, md5.Input = "The 128-bit hash of the 1990s, now broken: collisions take seconds. Shown to learn from, never to use.", "O(n)", "O(n)", Text
	md5.Example = &Input{Text: "abc"}
	md5.Run = func(in Input, _ trace.Recorder) (Output, error) {
		sum := cryptography.MD5([]byte(in.Text))
		return Output{Text: cryptography.Hex(sum[:])}, nil
	}
	md5.Call = func(in Input) (string, string) { return hexCall(fmt.Sprintf("cryptography.MD5([]byte(%q))", in.Text)) }

	mac := entry("hmac-sha256", "HMAC-SHA256", "security/cryptography/hmac.go", "HMACSHA256")
	mac.Summary, mac.Time, mac.Space, mac.Input, mac.Params = "Authenticate a message with a key: two nested SHA-256 hashes, safe where hash(key + message) is not.", "O(n)", "O(n)", Text, []string{"key"}
	mac.Example = &Input{Text: "what do ya want for nothing?", Key: "Jefe"}
	mac.Run = func(in Input, _ trace.Recorder) (Output, error) {
		sum := cryptography.HMACSHA256([]byte(in.Key), []byte(in.Text))
		return Output{Text: cryptography.Hex(sum[:])}, nil
	}
	mac.Call = func(in Input) (string, string) {
		return hexCall(fmt.Sprintf("cryptography.HMACSHA256([]byte(%q), []byte(%q))", in.Key, in.Text))
	}

	pbkdf2 := entry("pbkdf2", "PBKDF2", "security/cryptography/hmac.go", "PBKDF2")
	pbkdf2.Summary, pbkdf2.Time, pbkdf2.Space, pbkdf2.Input, pbkdf2.Params = "Stretch a password (the text) with a salt (the key) into a key, with many HMAC rounds to slow down guessing.", "O(iterations * length)", "O(length)", Text, []string{"key"}
	pbkdf2.Options = []Option{{"iterations", "iterations", 1000, 1, 100000}, {"length", "key bytes", 32, 1, 128}}
	pbkdf2.Example = &Input{Text: "password", Key: "salt", Options: map[string]int{"iterations": 1000, "length": 32}}
	pbkdf2.Run = func(in Input, _ trace.Recorder) (Output, error) {
		it, n := in.opt("iterations", 1000), in.opt("length", 32)
		if it < 1 || it > 100000 || n < 1 || n > 128 {
			return Output{}, inputError("iterations from 1 to 100000, length from 1 to 128")
		}
		return Output{Text: cryptography.Hex(cryptography.PBKDF2([]byte(in.Text), []byte(in.Key), it, n))}, nil
	}
	pbkdf2.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("cryptography.Hex(cryptography.PBKDF2([]byte(%q), []byte(%q), %d, %d))", in.Text, in.Key, in.opt("iterations", 1000), in.opt("length", 32))
	}

	return append([]Entry{sha, md5, mac, pbkdf2, aesEntry(), chachaEntry()}, publicKeyEntries()...)
}

func aesEntry() Entry {
	e := entry("aes-128", "AES-128", "security/cryptography/aes.go", "AES128")
	e.Summary, e.Time, e.Space, e.Input, e.Params = "Encrypt one 16-byte block (text, in hex) with a 16-byte key: ten rounds of SubBytes, ShiftRows, MixColumns, AddRoundKey.", "O(1)", "O(1)", Text, []string{"key"}
	e.Limits = "one block: 32 hex digits of text and of key"
	e.Example = &Input{Text: "00112233445566778899aabbccddeeff", Key: "000102030405060708090a0b0c0d0e0f"}
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		b, err := hexOf(in.Text, 16, "the block")
		if err != nil {
			return Output{}, err
		}
		k, err := hexOf(in.Key, 16, "the key")
		if err != nil {
			return Output{}, err
		}
		var key, block [16]byte
		copy(key[:], k)
		copy(block[:], b)
		out, steps := cryptography.TraceAES128(key, block)
		t := &TableOutput{Cols: 16}
		for i := 0; i < 16; i++ {
			t.ColLabels = append(t.ColLabels, fmt.Sprintf("r%dc%d", i%4, i/4))
		}
		for _, s := range steps {
			t.Rows++
			t.RowLabels = append(t.RowLabels, fmt.Sprintf("round %d %s", s.Round, s.Step))
			for i := 0; i < 16; i++ {
				t.Cells = append(t.Cells, int(s.State[i%4][i/4]))
			}
		}
		return Output{Text: cryptography.Hex(out[:]), Table: t}, nil
	}
	e.Call = func(in Input) (string, string) {
		return fmt.Sprintf("var key, block [16]byte\n\tk, _ := cryptography.HexDecode(%q)\n\tb, _ := cryptography.HexDecode(%q)\n\tcopy(key[:], k)\n\tcopy(block[:], b)\n\tsum := cryptography.AES128(key, block)\n\t", in.Key, in.Text),
			"cryptography.Hex(sum[:])"
	}
	return e
}

// parseChaCha reads the key as "KEY NONCE": 64 and 24 hex digits.
func parseChaCha(s string) (key [32]byte, nonce [12]byte, err error) {
	f := strings.Fields(s)
	if len(f) != 2 {
		return key, nonce, inputError("give the key as 64 hex digits, a space, and the nonce as 24")
	}
	k, err := hexOf(f[0], 32, "the key")
	if err != nil {
		return key, nonce, err
	}
	n, err := hexOf(f[1], 12, "the nonce")
	if err != nil {
		return key, nonce, err
	}
	copy(key[:], k)
	copy(nonce[:], n)
	return key, nonce, nil
}

func chachaEntry() Entry {
	e := entry("chacha20", "ChaCha20", "security/cryptography/chacha20.go", "ChaCha20")
	e.Summary, e.Time, e.Space, e.Input, e.Params = "A stream cipher of additions, XORs and rotations; the key field is \"KEY NONCE\" in hex. Shows the state after each double round.", "O(n)", "O(n)", Text, []string{"key"}
	e.Options = []Option{{"counter", "first block counter", 1, 0, 1 << 30}}
	e.Example = &Input{
		Text:    "Ladies and Gentlemen of the class of '99",
		Key:     "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f 000000000000004a00000000",
		Options: map[string]int{"counter": 1},
	}
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		key, nonce, err := parseChaCha(in.Key)
		if err != nil {
			return Output{}, err
		}
		ctr := uint32(in.opt("counter", 1))
		rounds := cryptography.ChaCha20Rounds(key, ctr, nonce)
		t := &TableOutput{Rows: len(rounds), Cols: 16}
		for i, s := range rounds {
			t.RowLabels = append(t.RowLabels, "double round "+strconv.Itoa(i))
			for _, w := range s {
				t.Cells = append(t.Cells, int(w))
			}
		}
		t.RowLabels[0] = "start"
		return Output{Text: cryptography.Hex(cryptography.ChaCha20(key, nonce, ctr, []byte(in.Text))), Table: t}, nil
	}
	e.Call = func(in Input) (string, string) {
		f := strings.Fields(in.Key)
		if len(f) != 2 {
			return "", ""
		}
		return fmt.Sprintf("var key [32]byte\n\tvar nonce [12]byte\n\tk, _ := cryptography.HexDecode(%q)\n\tn, _ := cryptography.HexDecode(%q)\n\tcopy(key[:], k)\n\tcopy(nonce[:], n)\n\t", f[0], f[1]),
			fmt.Sprintf("cryptography.Hex(cryptography.ChaCha20(key, nonce, %d, []byte(%q)))", in.opt("counter", 1), in.Text)
	}
	return e
}

func publicKeyEntries() []Entry {
	rsa := entry("rsa", "RSA", "security/cryptography/rsa.go", "GenerateRSA")
	rsa.Summary, rsa.Time, rsa.Space, rsa.Input = "Make a key from two seeded primes, then encrypt, decrypt, sign and verify a number with textbook RSA.", "O(bits^4) to make the key", "O(bits)", Number
	rsa.Options = []Option{{"bits", "modulus bits", 128, 64, 1024}, {"seed", "seed", 42, 0, 1 << 30}}
	rsa.Example = &Input{Values: []int{42}, Options: map[string]int{"bits": 128, "seed": 42}}
	rsa.Run = func(in Input, _ trace.Recorder) (Output, error) {
		v, err := one(in)
		if err != nil {
			return Output{}, err
		}
		bitLen := in.opt("bits", 128)
		if bitLen > 1024 {
			return Output{}, inputError("at most 1024 bits here")
		}
		k, err := cryptography.GenerateRSA(bitLen, int64(in.opt("seed", 42)))
		if err != nil {
			return Output{}, wrap(err)
		}
		m := big.NewInt(int64(v))
		c, err := k.Encrypt(m)
		if err != nil {
			return Output{}, wrap(err)
		}
		sig := k.Sign(m)
		return Output{Lines: []string{
			"p = " + k.P.String(), "q = " + k.Q.String(), "N = p*q = " + k.N.String(),
			"e = " + k.E.String(), "d = e^-1 mod (p-1)(q-1) = " + k.D.String(),
			"ciphertext m^e mod N = " + c.String(), "decrypted c^d mod N = " + k.Decrypt(c).String(),
			"signature m^d mod N = " + sig.String(), fmt.Sprintf("verifies: %v", k.Verify(m, sig)),
		}}, nil
	}

	dh := entry("diffie-hellman", "Diffie-Hellman", "security/cryptography/dh.go", "DiffieHellman")
	dh.Summary, dh.Time, dh.Space, dh.Input = "Two sides agree on a shared secret over an open line: each sends g^secret mod p.", "O(log p)", "O(1)", Simulation
	dh.Options = []Option{{"p", "prime p", 23, 3, 1 << 31}, {"g", "generator g", 5, 2, 1 << 31}, {"a", "Alice's secret", 6, 1, 1 << 30}, {"b", "Bob's secret", 15, 1, 1 << 30}}
	dh.Example = &Input{Options: map[string]int{"p": 23, "g": 5, "a": 6, "b": 15}}
	dh.Run = func(in Input, _ trace.Recorder) (Output, error) {
		x, err := cryptography.DiffieHellman(int64(in.opt("p", 23)), int64(in.opt("g", 5)), int64(in.opt("a", 6)), int64(in.opt("b", 15)))
		if err != nil {
			return Output{}, wrap(err)
		}
		return Output{Number: ptr(int(x.Shared)), Summary: map[string]float64{"A = g^a mod p": float64(x.A), "B = g^b mod p": float64(x.B), "shared": float64(x.Shared)}}, nil
	}
	dh.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("cryptography.DiffieHellman(%d, %d, %d, %d)", in.opt("p", 23), in.opt("g", 5), in.opt("a", 6), in.opt("b", 15))
	}

	elg := entry("elgamal", "ElGamal", "security/cryptography/elgamal.go", "ElGamalEncrypt")
	elg.Summary, elg.Time, elg.Space, elg.Input = "Encrypt a number with a public key y = g^x mod p and a one-time k, then decrypt it with the secret x.", "O(log p)", "O(1)", Simulation
	elg.Options = []Option{{"p", "prime p", 467, 3, 1 << 31}, {"g", "generator g", 2, 2, 1 << 31}, {"x", "secret x", 127, 1, 1 << 30}, {"m", "message m", 100, 1, 1 << 31}, {"k", "one-time k", 213, 1, 1 << 31}}
	elg.Example = &Input{Options: map[string]int{"p": 467, "g": 2, "x": 127, "m": 100, "k": 213}}
	elgArgs := func(in Input) (p, g, x, m, k int64) {
		return int64(in.opt("p", 467)), int64(in.opt("g", 2)), int64(in.opt("x", 127)), int64(in.opt("m", 100)), int64(in.opt("k", 213))
	}
	elg.Run = func(in Input, _ trace.Recorder) (Output, error) {
		p, g, x, m, k := elgArgs(in)
		if p < 3 || p > 1<<31 {
			return Output{}, inputError("p must be a prime from 3 to 2^31")
		}
		y := cryptography.ModPow(g, x, p)
		c, err := cryptography.ElGamalEncrypt(p, g, y, m, k)
		if err != nil {
			return Output{}, wrap(err)
		}
		d := cryptography.ElGamalDecrypt(p, x, c)
		return Output{Number: ptr(int(d)), Lines: []string{
			fmt.Sprintf("public key y = g^x mod p = %d", y),
			fmt.Sprintf("c1 = g^k mod p = %d", c.C1),
			fmt.Sprintf("c2 = m * y^k mod p = %d", c.C2),
			fmt.Sprintf("decrypted c2 / c1^x mod p = %d", d),
		}}, nil
	}
	elg.Call = func(in Input) (string, string) {
		p, g, x, m, k := elgArgs(in)
		return fmt.Sprintf("y := cryptography.ModPow(%d, %d, %d)\n\tc, err := cryptography.ElGamalEncrypt(%d, %d, y, %d, %d)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", g, x, p, p, g, m, k),
			fmt.Sprintf("c, cryptography.ElGamalDecrypt(%d, %d, c)", p, x)
	}

	return append([]Entry{rsa, dh, elg, curveEntry(), shamirEntry(), merkleEntry()}, toyEntries()...)
}

func curveEntry() Entry {
	e := entry("elliptic-curve", "Elliptic Curve", "security/cryptography/ec.go", "Curve")
	e.Summary, e.Time, e.Space, e.Input = "Points on y^2 = x^3 + ax + b mod p: list them, and compute k*G by double-and-add.", "O(p^2) to list, O(log k) to multiply", "O(p)", Simulation
	e.Options = []Option{{"a", "a", 2, 0, 1000}, {"b", "b", 2, 0, 1000}, {"p", "prime p", 17, 3, 1000}, {"x", "G's x", 5, 0, 1000}, {"y", "G's y", 1, 0, 1000}, {"k", "k", 7, 1, 1 << 30}}
	e.Example = &Input{Options: map[string]int{"a": 2, "b": 2, "p": 17, "x": 5, "y": 1, "k": 7}}
	e.Limits = "p at most 1000, so every point can be listed"
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		c := cryptography.Curve{A: int64(in.opt("a", 2)), B: int64(in.opt("b", 2)), P: int64(in.opt("p", 17))}
		if c.P < 3 || c.P > 1000 || !isSmallPrime(c.P) {
			return Output{}, inputError("p must be a prime from 3 to 1000")
		}
		if (4*c.A*c.A*c.A+27*c.B*c.B)%c.P == 0 {
			return Output{}, inputError("4a^3 + 27b^2 is 0 mod p: the curve is singular")
		}
		g := cryptography.ECPoint{X: int64(in.opt("x", 5)), Y: int64(in.opt("y", 1))}
		if !c.OnCurve(g) {
			return Output{}, inputError(fmt.Sprintf("G = %v is not on the curve", g))
		}
		k := int64(in.opt("k", 7))
		res, steps := c.ScalarMult(k, g)
		pts := c.Points()
		names := make([]string, len(pts))
		for i, p := range pts {
			names[i] = p.String()
		}
		lines := append([]string{fmt.Sprintf("%d points: %s", len(pts), strings.Join(names, " ")), fmt.Sprintf("G = %v has order %d", g, c.Order(g))}, steps...)
		lines = append(lines, fmt.Sprintf("%d*G = %v", k, res))
		return Output{Lines: lines}, nil
	}
	e.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("cryptography.Curve{A: %d, B: %d, P: %d}.ScalarMult(%d, cryptography.ECPoint{X: %d, Y: %d})",
			in.opt("a", 2), in.opt("b", 2), in.opt("p", 17), in.opt("k", 7), in.opt("x", 5), in.opt("y", 1))
	}
	return e
}

func isSmallPrime(n int64) bool {
	return big.NewInt(n).ProbablyPrime(0)
}

func shamirEntry() Entry {
	e := entry("shamir-secret-sharing", "Shamir Secret Sharing", "security/cryptography/shamir.go", "ShamirSplit")
	e.Summary, e.Time, e.Space, e.Input = "Split a secret into n shares so any k rebuild it and k-1 reveal nothing: points on a random polynomial.", "O(n k)", "O(n)", Simulation
	e.Options = []Option{{"secret", "secret", 1234, 0, 1<<31 - 2}, {"k", "shares needed", 3, 1, 10}, {"n", "shares", 5, 1, 20}, {"seed", "seed", 1, 0, 1 << 30}}
	e.Example = &Input{Options: map[string]int{"secret": 1234, "k": 3, "n": 5, "seed": 1}}
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		k, n := in.opt("k", 3), in.opt("n", 5)
		if n > 20 {
			return Output{}, inputError("at most 20 shares here")
		}
		shares, err := cryptography.ShamirSplit(int64(in.opt("secret", 1234)), k, n, int64(in.opt("seed", 1)))
		if err != nil {
			return Output{}, wrap(err)
		}
		lines := make([]string, 0, n+2)
		for _, s := range shares {
			lines = append(lines, fmt.Sprintf("share (%d, %d)", s.X, s.Y))
		}
		got, err := cryptography.ShamirCombine(shares[n-k:])
		if err != nil {
			return Output{}, wrap(err)
		}
		lines = append(lines, fmt.Sprintf("the last %d shares give %d", k, got))
		if k > 1 {
			less, _ := cryptography.ShamirCombine(shares[n-k+1:])
			lines = append(lines, fmt.Sprintf("one fewer gives %d, which says nothing", less))
		}
		return Output{Number: ptr(int(got)), Lines: lines}, nil
	}
	e.Call = func(in Input) (string, string) {
		k := in.opt("k", 3)
		return fmt.Sprintf("shares, err := cryptography.ShamirSplit(%d, %d, %d, %d)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", in.opt("secret", 1234), k, in.opt("n", 5), in.opt("seed", 1)),
			fmt.Sprintf("cryptography.ShamirCombine(shares[:%d])", k)
	}
	return e
}

func merkleBlocks(in Input) [][]byte {
	var blocks [][]byte
	for _, l := range strings.Split(in.Text, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			blocks = append(blocks, []byte(l))
		}
	}
	return blocks
}

func merkleEntry() Entry {
	e := entry("merkle-tree", "Merkle Tree", "security/cryptography/merkle.go", "MerkleTree")
	e.Summary, e.Time, e.Space, e.Input = "Hash blocks (one a line) pairwise up to one root, and prove a block is in the tree with log n hashes.", "O(n)", "O(n)", Text
	e.Options = []Option{{"leaf", "block to prove", 0, 0, 1000}}
	e.Example = &Input{Text: "alice pays bob 5\nbob pays carol 2\ncarol pays dave 1\ndave pays alice 3\nerin pays bob 4", Options: map[string]int{"leaf": 2}}
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		blocks := merkleBlocks(in)
		t, err := cryptography.NewMerkleTree(blocks)
		if err != nil {
			return Output{}, inputError("give the blocks, one a line")
		}
		leaf := in.opt("leaf", 0)
		proof, err := t.Proof(leaf)
		if err != nil {
			return Output{}, inputError(fmt.Sprintf("the block to prove is from 0 to %d", len(blocks)-1))
		}
		root := t.Root()
		lines := []string{fmt.Sprintf("proof for block %d (%q):", leaf, blocks[leaf])}
		for _, s := range proof {
			side := "right"
			if s.Left {
				side = "left"
			}
			lines = append(lines, fmt.Sprintf("  sibling on the %s: %s", side, cryptography.Hex(s.Hash[:4])))
		}
		ok := cryptography.VerifyMerkle(root, blocks[leaf], proof)
		return Output{
			Text: cryptography.Hex(root[:]), Lines: lines, Bool: &ok,
			Steps: []Step{{Op: Op{Name: "build"}, Result: cryptography.Hex(root[:4]), Snapshot: t.Snapshot()}},
		}, nil
	}
	e.Call = func(in Input) (string, string) {
		blocks := merkleBlocks(in)
		q := make([]string, len(blocks))
		for i, b := range blocks {
			q[i] = fmt.Sprintf("[]byte(%q)", b)
		}
		return fmt.Sprintf("t, err := cryptography.NewMerkleTree([][]byte{%s})\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\tsum := t.Root()\n\t", strings.Join(q, ", ")),
			"cryptography.Hex(sum[:])"
	}
	return e
}

// feistelKeys derives the round keys from one 16-bit key: key rotated left
// by 3 per round, plus the round number.
func feistelKeys(key, rounds int) []uint16 {
	keys := make([]uint16, rounds)
	for i := range keys {
		keys[i] = bits.RotateLeft16(uint16(key), 3*i) + uint16(i)
	}
	return keys
}

func uint16s(v []uint16) string {
	s := make([]string, len(v))
	for i, x := range v {
		s[i] = fmt.Sprintf("0x%04x", x)
	}
	return "[]uint16{" + strings.Join(s, ", ") + "}"
}

func cryptographyLfsrArgs(in Input) (seed uint32, taps []int, width, n int) {
	width, n = in.opt("width", 4), in.opt("n", 32)
	mask := in.opt("taps", 3)
	for b := 0; b < 20; b++ {
		if mask>>b&1 == 1 {
			taps = append(taps, b)
		}
	}
	return uint32(in.opt("seed", 1)), taps, width, n
}

func feistelNetworkEntry() Entry {
	feistel := entry("feistel-network", "Feistel Network", "security/cryptography/feistel.go", "Feistel")
	feistel.Summary, feistel.Time, feistel.Space, feistel.Input = "Split a 32-bit block in halves; each round swaps them and mixes one into the other. Decrypting is the same in reverse.", "O(rounds)", "O(rounds)", Simulation
	feistel.Options = []Option{{"block", "32-bit block", 0x12345678, 0, 1<<32 - 1}, {"key", "16-bit key", 0xbeef, 0, 1<<16 - 1}, {"rounds", "rounds", 8, 1, 32}}
	feistel.Example = &Input{Options: map[string]int{"block": 0x12345678, "key": 0xbeef, "rounds": 8}}
	feistel.Run = func(in Input, _ trace.Recorder) (Output, error) {
		block, key, rounds := in.opt("block", 0x12345678), in.opt("key", 0xbeef), in.opt("rounds", 8)
		if block < 0 || block > 1<<32-1 || key < 0 || key > 1<<16-1 || rounds < 1 || rounds > 32 {
			return Output{}, inputError("block from 0 to 2^32-1, key from 0 to 65535, rounds from 1 to 32")
		}
		keys := feistelKeys(key, rounds)
		out, steps := cryptography.Feistel(uint32(block), keys)
		back, _ := cryptography.FeistelDecrypt(out, keys)
		t := &TableOutput{Cols: 3, ColLabels: []string{"key", "left", "right"}}
		for i, s := range steps {
			t.Rows++
			t.RowLabels = append(t.RowLabels, "round "+strconv.Itoa(s.Round))
			t.Cells = append(t.Cells, int(keys[i]), int(s.Left), int(s.Right))
		}
		return Output{Number: ptr(int(out)), Table: t, Lines: []string{fmt.Sprintf("%08x encrypts to %08x and decrypts back to %08x", block, out, back)}}, nil
	}
	feistel.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("cryptography.Feistel(%d, %s)", in.opt("block", 0x12345678), uint16s(feistelKeys(in.opt("key", 0xbeef), in.opt("rounds", 8))))
	}
	return feistel
}

func lfsrEntry() Entry {
	lfsr := entry("lfsr", "Linear-Feedback Shift Register", "security/cryptography/lfsr.go", "LFSR")
	lfsr.Summary, lfsr.Time, lfsr.Space, lfsr.Input, lfsr.Traced = "Shift a register right and feed the XOR of its tapped bits back in at the top: a fast, predictable bit stream.", "O(n * taps)", "O(n)", Simulation, true
	lfsr.Options = []Option{{"seed", "seed", 1, 1, 1<<20 - 1}, {"width", "bits", 4, 2, 20}, {"taps", "tap mask (bit i taps bit i)", 3, 1, 1<<20 - 1}, {"n", "bits out", 32, 1, 1000}}
	lfsr.Example = &Input{Options: map[string]int{"seed": 1, "width": 4, "taps": 3, "n": 32}}
	lfsr.Run = func(in Input, r trace.Recorder) (Output, error) {
		seed, taps, width, n := cryptographyLfsrArgs(in)
		if width > 20 || n < 1 || n > 1000 {
			return Output{}, inputError("at most 20 bits wide and 1000 bits out")
		}
		out, err := cryptography.TraceLFSR(seed, taps, width, n, r)
		if err != nil {
			return Output{}, wrap(err)
		}
		period, _ := cryptography.LFSRPeriod(seed, taps, width)
		var b strings.Builder
		for _, x := range out {
			b.WriteByte(byte('0' + x))
		}
		return Output{Text: b.String(), Number: ptr(period), Lines: []string{fmt.Sprintf("taps %v, period %d of at most %d", taps, period, 1<<width-1)}}, nil
	}
	lfsr.Call = func(in Input) (string, string) {
		seed, taps, width, n := cryptographyLfsrArgs(in)
		return "", fmt.Sprintf("cryptography.LFSR(%d, %s, %d, %d)", seed, ints(taps), width, n)
	}
	return lfsr
}

func toyEntries() []Entry {
	return concat(
		[]Entry{
			feistelNetworkEntry(),
			lfsrEntry(),
		},
		encodingEntries(),
	)
}

func cryptographyEnc(id, name, fn, summary string, f func([]byte) string, ex string) Entry {
	e := entry(id, name, "security/cryptography/encoding.go", fn)
	e.Summary, e.Time, e.Space, e.Input, e.Example = summary, "O(n)", "O(n)", Text, &Input{Text: ex}
	e.Run = func(in Input, _ trace.Recorder) (Output, error) { return Output{Text: f([]byte(in.Text))}, nil }
	e.Call = func(in Input) (string, string) { return "", fmt.Sprintf("cryptography.%s([]byte(%q))", fn, in.Text) }
	return e
}

func constantTimeCompareEntry() Entry {
	ct := entry("constant-time-compare", "Constant-Time Compare", "security/cryptography/ctcompare.go", "ConstantTimeEqual")
	ct.Summary, ct.Time, ct.Space, ct.Input = "Compare two secrets looking at every byte, so the time taken leaks nothing about where they differ.", "O(n)", "O(1)", TwoTexts
	ct.Example = &Input{Text: "s3cret-token-42", Other: "s3cret-tokXn-42"}
	ct.Run = func(in Input, _ trace.Recorder) (Output, error) {
		a, b := []byte(in.Text), []byte(in.Other)
		eq := cryptography.ConstantTimeEqual(a, b)
		_, looked := cryptography.LeakyEqual(a, b)
		return Output{Bool: &eq, Lines: []string{
			fmt.Sprintf("an early-exit compare looked at %d byte(s): its timing tells an attacker how much of a guess is right", looked),
			fmt.Sprintf("the constant-time compare looked at all %d", len(a)),
		}}, nil
	}
	ct.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("cryptography.ConstantTimeEqual([]byte(%q), []byte(%q))", in.Text, in.Other)
	}
	return ct
}

func base64Entry() Entry {
	return cryptographyEnc("base64", "Base64", "Base64", "Write bytes as text: every 3 bytes become 4 of 64 safe characters. An encoding, not encryption.", cryptography.Base64, "Many hands make light work.")
}

func base32Entry() Entry {
	return cryptographyEnc("base32", "Base32", "Base32", "Write bytes as A-Z and 2-7, 5 bytes to 8 characters: case-insensitive, used in TOTP secrets.", cryptography.Base32, "foobar")
}

func hexEncodingEntry() Entry {
	return cryptographyEnc("hex-encoding", "Hex Encoding", "Hex", "Write each byte as two hex digits, 0-9 and a-f.", cryptography.Hex, "Hi!")
}

func encodingEntries() []Entry {
	return []Entry{
		base64Entry(),
		base32Entry(),
		hexEncodingEntry(),
		constantTimeCompareEntry(),
	}
}

// steppedOps is the entry for a structure whose operations each produce one
// or more steps (a hash table shows every probe). run applies one operation
// and returns its steps; fill is the operation each starting value goes
// through.
func steppedOps(e Entry, fill string, ops []string, example string, run func(in Input) func(op Op) ([]Step, error)) Entry {
	e.Input, e.Ops, e.Limits = Ops, ops, fmt.Sprintf("at most %d operations", maxOps)
	e.Example = &Input{Ops: ParseOps(example), Options: map[string]int{}}
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if len(in.Ops) > maxOps || len(in.Values) > maxOps {
			return Output{}, inputError(fmt.Sprintf("at most %d values and %d operations", maxOps, maxOps))
		}
		apply := run(in)
		var out Output
		for _, v := range in.Values {
			if _, err := apply(Op{Name: fill, Value: strconv.Itoa(v)}); err != nil {
				return Output{}, err
			}
		}
		start, _ := apply(Op{Name: "show"})
		out.Steps = append(out.Steps, Step{Op: Op{Name: "start"}, Result: fmt.Sprintf("%d values", len(in.Values)), Snapshot: start[0].Snapshot})
		for _, op := range in.Ops {
			steps, err := apply(op)
			if err != nil {
				st := Step{Op: op, Error: err.Error(), Snapshot: out.Steps[len(out.Steps)-1].Snapshot}
				out.Steps = append(out.Steps, st)
				continue
			}
			for i := range steps {
				steps[i].Op = op
			}
			out.Steps = append(out.Steps, steps...)
		}
		return out, nil
	}
	return e
}

func hashFunctionEntries() []Entry {
	fn := func(id, name, f, summary string, call func(text string, in Input) string, hash func(b []byte, in Input) uint64, opts []Option) Entry {
		e := entry(id, name, "security/hashing/functions.go", f)
		e.Summary, e.Time, e.Space, e.Input, e.Options = summary, "O(n)", "O(1)", Text, opts
		e.Example = &Input{Text: "hello world", Options: map[string]int{}}
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			if len(in.Text) > 1<<20 {
				return Output{}, inputError("at most 1 MiB of text")
			}
			var lines []string
			words := strings.Fields(in.Text)
			for _, w := range append([]string{in.Text}, words...) {
				h := hash([]byte(w), in)
				lines = append(lines, fmt.Sprintf("%q  0x%08x  %d", w, h, h))
				if len(words) <= 1 {
					break
				}
			}
			v := int(hash([]byte(in.Text), in) & 0x7fffffffffffffff)
			return Output{Number: &v, Lines: lines}, nil
		}
		e.Call = func(in Input) (string, string) { return "", call(in.Text, in) }
		return e
	}
	u := func(x uint32) uint64 { return uint64(x) }
	hashes := []Entry{
		fn("fnv1a", "FNV-1a (32 and 64 bit)", "FNV1a32", "For each byte, XOR it in and multiply by the FNV prime: simple, fast and well spread.",
			func(t string, _ Input) string { return fmt.Sprintf("hashing.FNV1a32([]byte(%q))", t) },
			func(b []byte, _ Input) uint64 { return u(hashing.FNV1a32(b)) }, nil),
		fn("djb2", "DJB2", "DJB2", "Daniel Bernstein's h = h*33 + byte, from 5381.",
			func(t string, _ Input) string { return fmt.Sprintf("hashing.DJB2([]byte(%q))", t) },
			func(b []byte, _ Input) uint64 { return u(hashing.DJB2(b)) }, nil),
		fn("sdbm", "SDBM", "SDBM", "The sdbm database's h = byte + (h << 6) + (h << 16) - h.",
			func(t string, _ Input) string { return fmt.Sprintf("hashing.SDBM([]byte(%q))", t) },
			func(b []byte, _ Input) uint64 { return u(hashing.SDBM(b)) }, nil),
		fn("murmur3", "MurmurHash3 (32 bit)", "Murmur3_32", "Mix four bytes at a time with multiplies and rotations, then avalanche: every input bit affects every output bit.",
			func(t string, in Input) string {
				return fmt.Sprintf("hashing.Murmur3_32([]byte(%q), %d)", t, in.opt("seed", 0))
			},
			func(b []byte, in Input) uint64 { return u(hashing.Murmur3_32(b, uint32(in.opt("seed", 0)))) },
			[]Option{{"seed", "seed", 0, 0, 1 << 30}}),
		fn("crc32", "CRC-32", "CRC32", "The remainder of the data as a polynomial over a fixed one, a table lookup a byte: catches accidental corruption.",
			func(t string, _ Input) string { return fmt.Sprintf("hashing.CRC32([]byte(%q))", t) },
			func(b []byte, _ Input) uint64 { return u(hashing.CRC32(b)) }, nil),
		fn("adler32", "Adler-32", "Adler32", "Two running sums modulo 65521, the second summing the first (zlib's checksum).",
			func(t string, _ Input) string { return fmt.Sprintf("hashing.Adler32([]byte(%q))", t) },
			func(b []byte, _ Input) uint64 { return u(hashing.Adler32(b)) }, nil),
		fn("polynomial-hash", "Polynomial Rolling Hash", "Polynomial", "sum of byte * base^position mod a prime: the hash a sliding window updates in O(1).",
			func(t string, in Input) string {
				return fmt.Sprintf("hashing.Polynomial([]byte(%q), %d, %d)", t, in.opt("base", 31), in.opt("mod", 1_000_000_007))
			},
			func(b []byte, in Input) uint64 {
				return hashing.Polynomial(b, uint64(in.opt("base", 31)), uint64(max(2, in.opt("mod", 1_000_000_007))))
			},
			[]Option{{"base", "base", 31, 2, 1000}, {"mod", "modulus", 1_000_000_007, 2, 2_000_000_000}}),
	}
	hashes[0].Summary += " The 64-bit hash is FNV1a64."

	funcs := []func([]byte) uint32{hashing.FNV1a32, hashing.DJB2, hashing.SDBM, func(b []byte) uint32 { return hashing.Murmur3_32(b, 0) }, hashing.CRC32}
	names := []string{"FNV-1a", "DJB2", "SDBM", "Murmur3", "CRC-32"}
	dist := entry("hash-distribution", "Hash Distribution", "security/hashing/distribution.go", "Distribution")
	dist.Summary, dist.Time, dist.Space, dist.Input = "How evenly a hash spreads keys over buckets: counts per bucket and a chi-square near buckets - 1 for a good hash.", "O(keys + buckets)", "O(buckets)", Text
	dist.Options = []Option{{"buckets", "buckets", 16, 1, 1024}, {"hash", "hash (0 FNV-1a, 1 DJB2, 2 SDBM, 3 Murmur3, 4 CRC-32)", 0, 0, 4}}
	var sample []string
	for i := 0; i < 200; i++ {
		sample = append(sample, "user"+strconv.Itoa(i))
	}
	dist.Example = &Input{Text: strings.Join(sample, " "), Options: map[string]int{"buckets": 16, "hash": 0}}
	dist.Run = func(in Input, _ trace.Recorder) (Output, error) {
		keys := strings.Fields(in.Text)
		if len(keys) == 0 || len(keys) > 100_000 {
			return Output{}, inputError("give 1 to 100000 keys, separated by spaces or new lines")
		}
		h, b := in.opt("hash", 0), in.opt("buckets", 16)
		if h < 0 || h >= len(funcs) || b < 1 || b > 1024 {
			return Output{}, inputError("hash is 0 to 4 and buckets 1 to 1024")
		}
		d := hashing.Distribution(keys, b, funcs[h])
		return Output{Table: &TableOutput{Rows: 1, Cols: b, Cells: d.Counts, RowLabels: []string{names[h]}, ColLabels: numberLabels(b, 0)},
			Summary: map[string]float64{"chi-square": d.ChiSquare, "expected chi-square (buckets - 1)": float64(b - 1), "empty buckets": float64(d.Empty), "fullest bucket": float64(d.Max)}}, nil
	}
	return append(hashes, dist, openEntry("linear-probing", "Linear Probing", hashing.Linear, "On a collision try the next slot, then the next: simple and cache friendly, but full runs grow together."),
		openEntry("quadratic-probing", "Quadratic Probing", hashing.Quadratic, "On a collision try 1, 4, 9, ... slots on: breaks up runs; the table stays at most half full."),
		openEntry("double-hashing", "Double Hashing", hashing.Double, "On a collision step by a second hash of the key, so keys that collide once usually part ways."),
		cuckooEntry(), robinHoodEntry(), consistentEntry())
}

func tableOpList() []string { return []string{"insert KEY", "search KEY", "delete KEY"} }

// slotRow draws a row of slots: label(i) and role(i) for each.
func slotRow(n int, label func(i int) string, role func(i int) string) trace.Snapshot {
	s := trace.Snapshot{Layout: "row", Nodes: make([]trace.Node, n)}
	for i := range s.Nodes {
		s.Nodes[i] = trace.Node{ID: i, Label: label(i), Role: role(i), Parent: -1}
	}
	return s
}

func openEntry(id, name string, probe hashing.Probe, summary string) Entry {
	e := entry(id, name, "security/hashing/openaddressing.go", "OpenTable")
	e.Summary, e.Time, e.Space = summary+" Deletes leave tombstones.", "O(1) average at a bounded load", "O(n)"
	e.Options = []Option{{"capacity", "starting slots (a prime is chosen)", 7, 3, 1000}}
	return steppedOps(e, "insert", tableOpList(), "insert 12, insert 19, insert 26, insert 5, search 26, delete 19, search 26, insert 33", func(in Input) func(Op) ([]Step, error) {
		t := hashing.NewOpenTable(in.opt("capacity", 7), probe)
		// frozen is the table as it is now, so probe steps can show the table
		// before an insert changed it.
		type slot struct{ key, state int }
		frozen := func() []slot {
			out := make([]slot, t.Cap())
			for i := range out {
				out[i].key, out[i].state = t.Slot(i)
			}
			return out
		}
		draw := func(slots []slot, marks map[int]string) trace.Snapshot {
			return slotRow(len(slots), func(i int) string {
				switch slots[i].state {
				case hashing.Used:
					return strconv.Itoa(slots[i].key)
				case hashing.Tombstone:
					return "deleted"
				}
				return ""
			}, func(i int) string {
				if r, ok := marks[i]; ok {
					return r
				}
				switch slots[i].state {
				case hashing.Free:
					return "free"
				case hashing.Tombstone:
					return "tombstone"
				}
				return ""
			})
		}
		snap := func(marks map[int]string) trace.Snapshot { return draw(frozen(), marks) }
		// probes turns a probe sequence into one step per slot looked at: the
		// table as it was before the operation, then as it is after.
		probes := func(before []slot, probed []int, last, result string) []Step {
			if len(before) != t.Cap() {
				before = frozen() // it resized: the probes are in the new table
			}
			var steps []Step
			for i, s := range probed {
				role, res, view := "probe", fmt.Sprintf("probe %d: slot %d", i+1, s), before
				if i == len(probed)-1 {
					role, res, view = last, result, frozen()
				}
				steps = append(steps, Step{Result: res, Snapshot: draw(view, map[int]string{s: role})})
			}
			if len(steps) == 0 {
				steps = append(steps, Step{Result: result, Snapshot: snap(nil)})
			}
			return steps
		}
		return func(op Op) ([]Step, error) {
			if op.Name == "show" {
				return []Step{{Snapshot: snap(nil)}}, nil
			}
			k, err := opInt(op)
			if err != nil {
				return nil, err
			}
			state := frozen()
			switch op.Name {
			case "insert":
				before := t.Cap()
				added, probed, err := t.Insert(k)
				if err != nil {
					return nil, wrap(err)
				}
				res := fmt.Sprintf("inserted %d after %d probe(s), load %.2f", k, len(probed), t.LoadFactor())
				if !added {
					res = fmt.Sprintf("%d is already there", k)
				}
				if t.Cap() != before {
					res += fmt.Sprintf(" (grew from %d to %d slots and rehashed)", before, t.Cap())
				}
				return probes(state, probed, "new", res), nil
			case "search":
				found, probed := t.Search(k)
				res, role := fmt.Sprintf("found %d after %d probe(s)", k, len(probed)), "hit"
				if !found {
					res, role = fmt.Sprintf("%d is not there: %d probe(s) ended at a free slot", k, len(probed)), "probe"
				}
				return probes(state, probed, role, res), nil
			case "delete":
				removed, probed := t.Delete(k)
				res, role := fmt.Sprintf("deleted %d, leaving a tombstone", k), "tombstone"
				if !removed {
					res, role = fmt.Sprintf("%d is not there", k), "probe"
				}
				return probes(state, probed, role, res), nil
			}
			return nil, unknownOp(op, "insert", "search", "delete")
		}
	})
}

func cuckooEntry() Entry {
	e := entry("cuckoo-hashing", "Cuckoo Hashing", "security/hashing/cuckoo.go", "Cuckoo")
	e.Summary, e.Time, e.Space = "Two tables, two hashes, every key in one of its two slots: an insert evicts the occupant to its other slot, and a cycle rehashes.", "O(1) lookups, O(1) amortized inserts", "O(n)"
	e.Options = []Option{{"capacity", "slots per table", 5, 2, 1000}}
	return steppedOps(e, "insert", tableOpList(), "insert 10, insert 20, insert 30, insert 40, insert 50, insert 60, search 30, delete 20", func(in Input) func(Op) ([]Step, error) {
		c := hashing.NewCuckoo(in.opt("capacity", 5))
		snap := func(mark map[[2]int]string) trace.Snapshot {
			s := trace.Snapshot{Layout: "tree", Nodes: []trace.Node{{ID: 0, Label: "tables", Role: "root", Parent: -1}}}
			for t := 0; t < 2; t++ {
				tid := len(s.Nodes)
				s.Nodes = append(s.Nodes, trace.Node{ID: tid, Label: "T" + strconv.Itoa(t+1), Role: "bucket", Parent: 0})
				s.Edges = append(s.Edges, trace.Edge{From: 0, To: tid})
				for i := 0; i < c.Cap(); i++ {
					id := len(s.Nodes)
					k, used := c.Slot(t, i)
					n := trace.Node{ID: id, Label: "", Role: "free", Parent: tid}
					if used {
						n.Label, n.Role = strconv.Itoa(k), ""
					}
					if r, ok := mark[[2]int{t, i}]; ok {
						n.Role = r
					}
					s.Nodes = append(s.Nodes, n)
					s.Edges = append(s.Edges, trace.Edge{From: tid, To: id, Label: strconv.Itoa(i)})
				}
			}
			return s
		}
		return func(op Op) ([]Step, error) {
			if op.Name == "show" {
				return []Step{{Snapshot: snap(nil)}}, nil
			}
			k, err := opInt(op)
			if err != nil {
				return nil, err
			}
			switch op.Name {
			case "insert":
				rehashes := c.Rehashes
				added, kicks := c.Insert(k)
				if !added {
					return []Step{{Result: fmt.Sprintf("%d is already there", k), Snapshot: snap(nil)}}, nil
				}
				var steps []Step
				for _, kk := range kicks {
					res := fmt.Sprintf("put %d in T%d slot %d", kk.Key, kk.Table+1, kk.Slot)
					if kk.EvictedOK {
						res += fmt.Sprintf(", evicting %d to its other table", kk.Evicted)
					}
					steps = append(steps, Step{Result: res, Snapshot: snap(map[[2]int]string{{kk.Table, kk.Slot}: "new"})})
				}
				if c.Rehashes != rehashes {
					steps = append(steps, Step{Result: fmt.Sprintf("a cycle: grew to %d slots per table and rehashed every key", c.Cap()), Snapshot: snap(nil)})
				}
				return steps, nil
			case "search":
				return []Step{{Result: fmt.Sprintf("contains %d: %v (at most two slots read)", k, c.Contains(k)), Snapshot: snap(nil)}}, nil
			case "delete":
				return []Step{{Result: fmt.Sprintf("deleted %d: %v", k, c.Delete(k)), Snapshot: snap(nil)}}, nil
			}
			return nil, unknownOp(op, "insert", "search", "delete")
		}
	})
}

func robinHoodEntry() Entry {
	e := entry("robin-hood-hashing", "Robin Hood Hashing", "security/hashing/robinhood.go", "RobinHood")
	e.Summary, e.Time, e.Space = "Linear probing where a key far from home takes the slot of a key nearer home, keeping every probe distance close to the average.", "O(1) expected", "O(n)"
	e.Options = []Option{{"capacity", "slots", 8, 4, 1000}}
	return steppedOps(e, "insert", tableOpList(), "insert 3, insert 11, insert 19, insert 4, insert 27, search 27, delete 11, search 27", func(in Input) func(Op) ([]Step, error) {
		t := hashing.NewRobinHood(in.opt("capacity", 8))
		snap := func(marks map[int]string) trace.Snapshot {
			return slotRow(t.Cap(), func(i int) string {
				if k, d := t.Slot(i); d >= 0 {
					return fmt.Sprintf("%d (d=%d)", k, d)
				}
				return ""
			}, func(i int) string {
				if r, ok := marks[i]; ok {
					return r
				}
				if _, d := t.Slot(i); d < 0 {
					return "free"
				}
				return ""
			})
		}
		steps := func(probed []int, last, result string) []Step {
			var out []Step
			for i, s := range probed {
				role, res := "probe", fmt.Sprintf("probe %d: slot %d", i+1, s)
				if i == len(probed)-1 {
					role, res = last, result
				}
				out = append(out, Step{Result: res, Snapshot: snap(map[int]string{s: role})})
			}
			return out
		}
		return func(op Op) ([]Step, error) {
			if op.Name == "show" {
				return []Step{{Snapshot: snap(nil)}}, nil
			}
			k, err := opInt(op)
			if err != nil {
				return nil, err
			}
			switch op.Name {
			case "insert":
				added, probed := t.Insert(k)
				res := fmt.Sprintf("inserted %d; the longest probe distance is %d", k, t.MaxDistance())
				if !added {
					res = fmt.Sprintf("%d is already there", k)
				}
				return steps(probed, "new", res), nil
			case "search":
				ok, probed := t.Search(k)
				role := "hit"
				if !ok {
					role = "probe"
				}
				return steps(probed, role, fmt.Sprintf("contains %d: %v after %d probe(s)", k, ok, len(probed))), nil
			case "delete":
				ok, probed := t.Delete(k)
				return steps(probed, "probe", fmt.Sprintf("deleted %d: %v; the keys after it shifted back, no tombstone", k, ok)), nil
			}
			return nil, unknownOp(op, "insert", "search", "delete")
		}
	})
}

func consistentEntry() Entry {
	e := entry("consistent-hashing", "Consistent Hashing", "security/hashing/consistent.go", "Ring")
	e.Summary, e.Time, e.Space = "Nodes and keys on one hash ring, each key owned by the next node clockwise: adding or removing a node moves only about 1/n of the keys.", "O(log n) per lookup", "O(n * replicas)"
	e.Options = []Option{{"replicas", "virtual nodes per node", 20, 1, 500}}
	return steppedOps(e, "add-node", []string{"add-node NAME", "remove-node NAME", "locate KEY"}, "add-node A, add-node B, add-node C, locate alice, locate bob, locate carol, locate dave, locate erin, add-node D, remove-node A", func(in Input) func(Op) ([]Step, error) {
		g := hashing.NewRing(in.opt("replicas", 20))
		var keys []string
		snap := func(highlight string) trace.Snapshot {
			s := trace.Snapshot{Layout: "tree", Nodes: []trace.Node{{ID: 0, Label: "ring", Role: "root", Parent: -1}}}
			ids := map[string]int{}
			for _, n := range g.Nodes() {
				id := len(s.Nodes)
				ids[n] = id
				role := ""
				if n == highlight {
					role = "new"
				}
				s.Nodes = append(s.Nodes, trace.Node{ID: id, Label: n, Role: role, Parent: 0})
				s.Edges = append(s.Edges, trace.Edge{From: 0, To: id})
			}
			for _, k := range keys {
				n, ok := g.Locate(k)
				if !ok {
					continue
				}
				id := len(s.Nodes)
				role := ""
				if k == highlight {
					role = "hit"
				}
				s.Nodes = append(s.Nodes, trace.Node{ID: id, Label: k, Role: role, Parent: ids[n]})
				s.Edges = append(s.Edges, trace.Edge{From: ids[n], To: id})
			}
			return s
		}
		return func(op Op) ([]Step, error) {
			name := strings.TrimSpace(op.Value)
			if op.Name != "show" && name == "" {
				return nil, inputError(op.Name + " needs a name")
			}
			before := g.Assign(keys)
			switch op.Name {
			case "show":
				return []Step{{Snapshot: snap("")}}, nil
			case "add-node", "remove-node":
				if op.Name == "add-node" {
					g.Add(name)
				} else {
					g.Remove(name)
				}
				moved := hashing.Moved(before, g.Assign(keys))
				return []Step{{Result: fmt.Sprintf("%s %s: %d of %d keys moved", strings.TrimSuffix(op.Name, "-node"), name, moved, len(keys)), Snapshot: snap(name)}}, nil
			case "locate":
				if !slices.Contains(keys, name) && len(keys) < maxOps {
					keys = append(keys, name)
				}
				n, ok := g.Locate(name)
				if !ok {
					return []Step{{Result: "the ring has no nodes", Snapshot: snap(name)}}, nil
				}
				return []Step{{Result: fmt.Sprintf("%s belongs to node %s", name, n), Snapshot: snap(name)}}, nil
			}
			return nil, unknownOp(op, "add-node", "remove-node", "locate")
		}
	})
}
