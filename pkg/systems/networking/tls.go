package networking

// TLSMessage is one message of a TLS 1.3 handshake.
type TLSMessage struct {
	Flight int    `json:"flight"` // 1-based; each flight is half a round trip
	From   string `json:"from"`   // "client" or "server"
	Name   string `json:"name"`
	// Encrypted is true for messages sent under the handshake keys (shown
	// in braces in RFC 8446).
	Encrypted bool   `json:"encrypted"`
	Purpose   string `json:"purpose"`
	// Uses names the catalog entries that show the cryptography inside it.
	Uses []string `json:"uses,omitempty"`
}

// TLS13 lists the messages of a TLS 1.3 handshake (RFC 8446) with what each
// is for; no cryptography is performed. A full handshake takes one round
// trip: the client guesses the key exchange and sends its key share in
// ClientHello, so after ServerHello both sides derive the handshake keys
// and everything else is encrypted. With resume, the client presents a
// pre-shared key from an earlier session and may send early (0-RTT) data
// in its first flight, and the server skips the certificate.
//
// Time: O(1). Space: O(1).
func TLS13(resume bool) []TLSMessage {
	if resume {
		return []TLSMessage{
			{1, "client", "ClientHello + key_share + pre_shared_key + early_data", false, "resume a session with a ticket's key, and offer a fresh key share for forward secrecy", []string{"elliptic-curve", "hmac-sha256"}},
			{1, "client", "(early data: GET /)", true, "0-RTT: the request rides in the first flight, under keys from the ticket; replayable, so only for safe requests", []string{"aes-128"}},
			{2, "server", "ServerHello + key_share + pre_shared_key", false, "accept the ticket; both derive the handshake keys", []string{"elliptic-curve", "pbkdf2"}},
			{2, "server", "{EncryptedExtensions}", true, "accept early data", nil},
			{2, "server", "{Finished}", true, "a MAC over the whole transcript", []string{"hmac-sha256"}},
			{3, "client", "{EndOfEarlyData}", true, "the end of 0-RTT data", nil},
			{3, "client", "{Finished}", true, "a MAC over the transcript; application keys from here", []string{"hmac-sha256"}},
		}
	}
	return []TLSMessage{
		{1, "client", "ClientHello + key_share", false, "random nonce, cipher suites (TLS_AES_128_GCM_SHA256 ...), and an ephemeral X25519 public key", []string{"elliptic-curve", "diffie-hellman"}},
		{2, "server", "ServerHello + key_share", false, "the chosen suite and the server's ephemeral public key: both now compute the shared secret and derive handshake keys", []string{"elliptic-curve", "diffie-hellman", "sha-256"}},
		{2, "server", "{EncryptedExtensions}", true, "the rest of the negotiation (ALPN: h2 or http/1.1), now hidden", nil},
		{2, "server", "{Certificate}", true, "the server's certificate chain", nil},
		{2, "server", "{CertificateVerify}", true, "a signature over the transcript with the certificate's key: proves the server holds it", []string{"rsa", "sha-256"}},
		{2, "server", "{Finished}", true, "a MAC over the transcript with a key from the handshake secret", []string{"hmac-sha256"}},
		{2, "server", "[application data]", true, "the server may send data already", []string{"aes-128", "chacha20"}},
		{3, "client", "{Finished}", true, "the client's MAC over the transcript; the handshake is done after one round trip", []string{"hmac-sha256"}},
		{3, "client", "[application data: GET /]", true, "the request, under the application keys", []string{"aes-128", "chacha20"}},
	}
}
