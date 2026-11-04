package cryptography_test

import (
	"bytes"
	"crypto/aes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math/big"
	"math/rand"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/security/cryptography"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func unhex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// inputs are messages of every length around the 55/56/64-byte padding edges.
func inputs() [][]byte {
	rng := rand.New(rand.NewSource(1))
	out := [][]byte{nil, []byte("abc")}
	for n := 1; n < 200; n += 1 + n/20 {
		b := make([]byte, n)
		rng.Read(b)
		out = append(out, b)
	}
	return out
}

func TestHashesMatchTheStandardLibrary(t *testing.T) {
	abc := cryptography.SHA256([]byte("abc"))
	if hex.EncodeToString(abc[:]) != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Errorf("SHA256(abc) = %x", abc)
	}
	if m := cryptography.MD5([]byte("abc")); hex.EncodeToString(m[:]) != "900150983cd24fb0d6963f7d28e17f72" {
		t.Errorf("MD5(abc) = %x", m)
	}
	for _, msg := range inputs() {
		if cryptography.SHA256(msg) != sha256.Sum256(msg) {
			t.Fatalf("SHA256 differs for %d bytes", len(msg))
		}
		if cryptography.MD5(msg) != md5.Sum(msg) {
			t.Fatalf("MD5 differs for %d bytes", len(msg))
		}
		for _, key := range [][]byte{nil, []byte("key"), bytes.Repeat([]byte{7}, 100)} {
			m := hmac.New(sha256.New, key)
			m.Write(msg)
			if got := cryptography.HMACSHA256(key, msg); !bytes.Equal(got[:], m.Sum(nil)) {
				t.Fatalf("HMAC differs for %d-byte key, %d-byte message", len(key), len(msg))
			}
		}
	}
	r := cryptography.SHA256Rounds([]byte("abc"))
	if r[63][0] != 0x506e3058 { // FIPS 180-2 appendix B.1, a after round 63
		t.Errorf("round 63 a = %08x", r[63][0])
	}
}

func TestHMACRFC4231(t *testing.T) {
	got := cryptography.HMACSHA256(bytes.Repeat([]byte{0x0b}, 20), []byte("Hi There"))
	if hex.EncodeToString(got[:]) != "b0344c61d8db38535ca8afceaf0bf12b881dc200c9833da726e9376c2e32cff7" {
		t.Errorf("HMAC case 1 = %x", got)
	}
	got = cryptography.HMACSHA256([]byte("Jefe"), []byte("what do ya want for nothing?"))
	if hex.EncodeToString(got[:]) != "5bdcc146bf60754e6a042426089575c75a003f089d2739839dec58b964ec3843" {
		t.Errorf("HMAC case 2 = %x", got)
	}
}

// pbkdf2Ref is PBKDF2 written over crypto/hmac, to check against.
func pbkdf2Ref(pw, salt []byte, iter, n int) []byte {
	var out []byte
	for block := uint32(1); len(out) < n; block++ {
		m := hmac.New(sha256.New, pw)
		m.Write(binary.BigEndian.AppendUint32(append([]byte(nil), salt...), block))
		u := m.Sum(nil)
		t := append([]byte(nil), u...)
		for i := 1; i < iter; i++ {
			m = hmac.New(sha256.New, pw)
			m.Write(u)
			u = m.Sum(nil)
			for j := range t {
				t[j] ^= u[j]
			}
		}
		out = append(out, t...)
	}
	return out[:n]
}

func TestPBKDF2(t *testing.T) {
	got := cryptography.PBKDF2([]byte("password"), []byte("salt"), 1, 32)
	if hex.EncodeToString(got) != "120fb6cffcf8b32c43e7225256c4f837a86548c92ccc35480805987cb70be17b" {
		t.Errorf("PBKDF2 c=1 = %x", got)
	}
	for _, c := range []struct{ iter, n int }{{2, 32}, {100, 50}, {3, 7}} {
		if got, want := cryptography.PBKDF2([]byte("pw"), []byte("NaCl"), c.iter, c.n), pbkdf2Ref([]byte("pw"), []byte("NaCl"), c.iter, c.n); !bytes.Equal(got, want) {
			t.Errorf("PBKDF2 %v = %x, want %x", c, got, want)
		}
	}
}

func TestAES(t *testing.T) {
	var key, pt [16]byte
	copy(key[:], unhex(t, "000102030405060708090a0b0c0d0e0f"))
	copy(pt[:], unhex(t, "00112233445566778899aabbccddeeff"))
	ct, steps := cryptography.TraceAES128(key, pt)
	if hex.EncodeToString(ct[:]) != "69c4e0d86a7b0430d8cdb78070b4c55a" {
		t.Errorf("FIPS-197 C.1 = %x", ct)
	}
	if len(steps) != 40 || steps[len(steps)-1].Round != 10 {
		t.Errorf("%d steps", len(steps))
	}
	if cryptography.AES128Decrypt(key, ct) != pt {
		t.Error("decrypt failed")
	}
	rng := rand.New(rand.NewSource(2))
	for i := 0; i < 200; i++ {
		rng.Read(key[:])
		rng.Read(pt[:])
		c, _ := aes.NewCipher(key[:])
		var want [16]byte
		c.Encrypt(want[:], pt[:])
		if got := cryptography.AES128(key, pt); got != want {
			t.Fatalf("AES differs from crypto/aes for key %x", key)
		}
		if cryptography.AES128Decrypt(key, want) != pt {
			t.Fatal("decrypt differs")
		}
	}
}

func TestChaCha20RFC8439(t *testing.T) {
	a, b, c, d := cryptography.QuarterRound(0x11111111, 0x01020304, 0x9b8d6f43, 0x01234567)
	if a != 0xea2a92f4 || b != 0xcb1cf8ce || c != 0x4581472e || d != 0x5881c4bb {
		t.Errorf("quarter round = %x %x %x %x", a, b, c, d)
	}
	var key [32]byte
	for i := range key {
		key[i] = byte(i)
	}
	var nonce [12]byte
	copy(nonce[:], unhex(t, "000000090000004a00000000"))
	if blk := cryptography.ChaCha20Block(key, 1, nonce); hex.EncodeToString(blk[:16]) != "10f1e7e4d13b5915500fdd1fa32071c4" {
		t.Errorf("block = %x", blk[:16])
	}
	copy(nonce[:], unhex(t, "000000000000004a00000000"))
	msg := []byte("Ladies and Gentlemen of the class of '99: If I could offer you only one tip for the future, sunscreen would be it.")
	ct := cryptography.ChaCha20(key, nonce, 1, msg)
	if hex.EncodeToString(ct[:16]) != "6e2e359a2568f98041ba0728dd0d6981" || hex.EncodeToString(ct[len(ct)-2:]) != "874d" {
		t.Errorf("ciphertext = %x", ct)
	}
	if !bytes.Equal(cryptography.ChaCha20(key, nonce, 1, ct), msg) {
		t.Error("round trip failed")
	}
}

func TestRSA(t *testing.T) {
	k, err := cryptography.GenerateRSA(512, 42)
	if err != nil {
		t.Fatal(err)
	}
	if k.N.BitLen() != 512 {
		t.Errorf("N has %d bits", k.N.BitLen())
	}
	k2, _ := cryptography.GenerateRSA(512, 42)
	if k2.N.Cmp(k.N) != 0 {
		t.Error("the same seed gave a different key")
	}
	for _, v := range []int64{0, 1, 2, 42, 1 << 40} {
		m := big.NewInt(v)
		c, err := k.Encrypt(m)
		if err != nil || k.Decrypt(c).Cmp(m) != 0 {
			t.Errorf("decrypt(encrypt(%d)) failed: %v", v, err)
		}
		s := k.Sign(m)
		if !k.Verify(m, s) || k.Verify(new(big.Int).Add(m, big.NewInt(1)), s) {
			t.Errorf("sign/verify of %d", v)
		}
	}
	if _, err := k.Encrypt(k.N); !errors.Is(err, cryptography.ErrInvalid) {
		t.Errorf("m = N: err = %v", err)
	}
	if _, err := cryptography.GenerateRSA(8, 1); !errors.Is(err, cryptography.ErrInvalid) {
		t.Error("tiny key accepted")
	}
}

func TestDiffieHellmanAndElGamal(t *testing.T) {
	ex, err := cryptography.DiffieHellman(23, 5, 6, 15)
	if err != nil || ex != (cryptography.DHExchange{A: 8, B: 19, Shared: 2}) {
		t.Errorf("DH = %+v, %v", ex, err)
	}
	if _, err := cryptography.DiffieHellman(24, 5, 6, 15); !errors.Is(err, cryptography.ErrInvalid) {
		t.Error("composite p accepted")
	}
	const p, g, x = 467, 2, 127
	y := cryptography.ModPow(g, x, p)
	for m := int64(1); m < p; m += 37 {
		c, err := cryptography.ElGamalEncrypt(p, g, y, m, 213)
		if err != nil || cryptography.ElGamalDecrypt(p, x, c) != m {
			t.Errorf("ElGamal %d: %+v %v", m, c, err)
		}
	}
}

func TestEllipticCurve(t *testing.T) {
	c := cryptography.Curve{A: 2, B: 2, P: 17}
	g := cryptography.ECPoint{X: 5, Y: 1}
	if d := c.Double(g); d != (cryptography.ECPoint{X: 6, Y: 3}) {
		t.Errorf("2G = %v", d)
	}
	if n := c.Order(g); n != 19 {
		t.Errorf("order = %d", n)
	}
	if pts := c.Points(); len(pts) != 19 {
		t.Errorf("%d points", len(pts))
	}
	q := cryptography.ECPoint{Inf: true}
	for k := int64(1); k <= 20; k++ {
		q = c.Add(q, g)
		got, steps := c.ScalarMult(k, g)
		if got != q || !c.OnCurve(got) || len(steps) == 0 {
			t.Errorf("%dG = %v, want %v", k, got, q)
		}
	}
}

func TestShamir(t *testing.T) {
	const secret = 123456789
	shares, err := cryptography.ShamirSplit(secret, 3, 6, 7)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 6; i++ {
		for j := i + 1; j < 6; j++ {
			for k := j + 1; k < 6; k++ {
				if got, err := cryptography.ShamirCombine([]cryptography.Share{shares[i], shares[j], shares[k]}); err != nil || got != secret {
					t.Errorf("shares %d %d %d give %d, %v", i, j, k, got, err)
				}
			}
		}
	}
	if got, _ := cryptography.ShamirCombine(shares[:2]); got == secret {
		t.Error("two shares were enough")
	}
	if _, err := cryptography.ShamirCombine([]cryptography.Share{shares[0], shares[0]}); !errors.Is(err, cryptography.ErrInvalid) {
		t.Error("repeated share accepted")
	}
}

func TestMerkle(t *testing.T) {
	for n := 1; n <= 9; n++ {
		var blocks [][]byte
		for i := 0; i < n; i++ {
			blocks = append(blocks, []byte{byte('a' + i)})
		}
		tree, err := cryptography.NewMerkleTree(blocks)
		if err != nil {
			t.Fatal(err)
		}
		for i, b := range blocks {
			proof, err := tree.Proof(i)
			if err != nil || !cryptography.VerifyMerkle(tree.Root(), b, proof) {
				t.Fatalf("n=%d: proof %d does not verify", n, i)
			}
			if cryptography.VerifyMerkle(tree.Root(), []byte("tampered"), proof) {
				t.Fatalf("n=%d: tampered block verified", n)
			}
			if len(proof) > 0 {
				proof[0].Hash[0] ^= 1
				if cryptography.VerifyMerkle(tree.Root(), b, proof) {
					t.Fatalf("n=%d: tampered proof verified", n)
				}
			}
		}
		if s := tree.Snapshot(); s.Nodes[0].Role != "root" && n > 1 {
			t.Errorf("snapshot root = %+v", s.Nodes[0])
		}
	}
	if _, err := cryptography.NewMerkleTree(nil); !errors.Is(err, cryptography.ErrInvalid) {
		t.Error("empty tree accepted")
	}
}

func TestFeistelAndLFSR(t *testing.T) {
	keys := []uint16{0x1234, 0xbeef, 0x0f0f, 0x7777}
	for _, block := range []uint32{0, 1, 0xdeadbeef, 0xffffffff} {
		c, rounds := cryptography.Feistel(block, keys)
		if p, _ := cryptography.FeistelDecrypt(c, keys); p != block || len(rounds) != 4 {
			t.Errorf("Feistel round trip of %x = %x", block, p)
		}
	}
	if n, err := cryptography.LFSRPeriod(1, []int{0, 1}, 4); err != nil || n != 15 {
		t.Errorf("4-bit period = %d, %v", n, err)
	}
	if n, _ := cryptography.LFSRPeriod(0xACE1, []int{0, 2, 3, 5}, 16); n != 65535 {
		t.Errorf("16-bit period = %d", n)
	}
	var log trace.Log
	bits, err := cryptography.TraceLFSR(1, []int{0, 1}, 4, 30, &log)
	if err != nil || len(bits) != 30 || log.Count(trace.Set) != 30 || !equalInts(bits[:15], bits[15:]) {
		t.Errorf("LFSR bits %v, %v", bits, err)
	}
	if _, err := cryptography.LFSR(0, []int{0}, 4, 3); !errors.Is(err, cryptography.ErrInvalid) {
		t.Error("zero seed accepted")
	}
}

func equalInts(a, b []int) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return len(a) == len(b)
}

func TestEncodings(t *testing.T) {
	for _, msg := range inputs() {
		if got := cryptography.Base64(msg); got != base64.StdEncoding.EncodeToString(msg) {
			t.Fatalf("Base64 = %q", got)
		}
		if got := cryptography.Base32(msg); got != base32.StdEncoding.EncodeToString(msg) {
			t.Fatalf("Base32 = %q", got)
		}
		if got := cryptography.Hex(msg); got != hex.EncodeToString(msg) {
			t.Fatalf("Hex = %q", got)
		}
		b64, err1 := cryptography.Base64Decode(cryptography.Base64(msg))
		b32, err2 := cryptography.Base32Decode(cryptography.Base32(msg))
		h, err3 := cryptography.HexDecode(cryptography.Hex(msg))
		if err := errors.Join(err1, err2, err3); err != nil || !bytes.Equal(b64, msg) || !bytes.Equal(b32, msg) || !bytes.Equal(h, msg) {
			t.Fatalf("decode of %d bytes: %v", len(msg), err)
		}
	}
	for _, bad := range []string{"a!", "QQ=x"} {
		if _, err := cryptography.Base64Decode(bad); !errors.Is(err, cryptography.ErrInvalid) {
			t.Errorf("Base64Decode(%q) err = %v", bad, err)
		}
	}
	if _, err := cryptography.HexDecode("abc"); !errors.Is(err, cryptography.ErrInvalid) {
		t.Error("odd hex accepted")
	}
}

func TestConstantTimeEqual(t *testing.T) {
	a, b := []byte("secret-token"), []byte("secret-tokeN")
	if !cryptography.ConstantTimeEqual(a, a) || cryptography.ConstantTimeEqual(a, b) || cryptography.ConstantTimeEqual(a, a[:3]) {
		t.Error("ConstantTimeEqual wrong")
	}
	if eq, n := cryptography.LeakyEqual(a, []byte("sXcret-token")); eq || n != 2 {
		t.Errorf("LeakyEqual = %v, %d", eq, n)
	}
}
