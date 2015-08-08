package jeeves

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"

	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/codykrieger/jeeves/ask"
)

var certURLStringToBytesMap map[string][]byte

func init() {
	certURLStringToBytesMap = make(map[string][]byte)
}

type Jeeves struct {
	skillsByEndpoint map[string]*Skill
}

type Skill struct {
	Name            string
	Endpoint        string
	ApplicationID   string
	Handler         func(*Skill, *ask.Request) *ask.Response
	internalHandler func(*requestContext, *ask.Request)
}

type requestContext struct {
	writer http.ResponseWriter
	req    *http.Request
	skill  *Skill
}

func New() *Jeeves {
	j := Jeeves{
		skillsByEndpoint: make(map[string]*Skill),
	}
	return &j
}

func (j *Jeeves) RegisterSkill(skill *Skill) {
	// FIXME: Validate Endpoint.
	j.skillsByEndpoint[skill.Endpoint] = skill

	skill.internalHandler = func(ctx *requestContext, req *ask.Request) {
		resp := skill.Handler(skill, req)
		bytes, _ := resp.Bytes()

		ctx.writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		ctx.writer.Write(bytes)
	}
}

func (j *Jeeves) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	skill := j.skillsByEndpoint[r.URL.Path]
	if skill == nil {
		http.NotFound(w, r)
		return
	}

	ctx := &requestContext{
		writer: w,
		req:    r,
		skill:  skill,
	}

	ctx.process()
}

func (ctx *requestContext) err(code int, theError error) {
	log.Printf("Error: %v; returning HTTP status code %v\n", theError, code)
	http.Error(ctx.writer, http.StatusText(code), code)
}

func (ctx *requestContext) process() {
	req, err := ask.NewRequestFromJSON(ctx.req.Body)
	if err != nil {
		ctx.err(400, err)
		return
	}

	if !ctx.validateRequestSignature(req) ||
		!ctx.validateTimestamp(req) ||
		!ctx.validateApplicationID(req) ||
		!ctx.validateRequestType(req) {
		return
	}

	ctx.skill.internalHandler(ctx, req)
}

func (ctx *requestContext) validateRequestSignature(req *ask.Request) bool {
	certChainURL := ctx.req.Header.Get("SignatureCertChainUrl")
	if !ctx.validateCertChainURL(certChainURL) {
		return false
	}

	certBytes, err := readCertAtURL(certChainURL)
	if err != nil {
		ctx.err(400, fmt.Errorf("Unable to fetch/parse certificate at given URL (%v): %v", certChainURL, err))
		return false
	}

	// FIXME: We need to check the additional blocks in the PEM structure to
	// ensure the Amazon signing certificate (the first certificate in the PEM
	// structure) has a valid chain of trust up to a trusted root CA (see
	// https://developer.amazon.com/public/solutions/alexa/alexa-skills-kit/docs/developing-an-alexa-skill-as-a-web-service#Checking%20the%20Signature%20of%20the%20Request).
	// This will at least be sufficient for development, though.
	pemBlock, _ := pem.Decode(certBytes)
	if err != nil {
		ctx.err(400, fmt.Errorf("Failed to parse certificate PEM: %v", err))
		return false
	}

	cert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		ctx.err(400, fmt.Errorf("Failed to parse x509 certificate: %v", err))
		return false
	}

	now := time.Now().Unix()
	if now < cert.NotBefore.Unix() || now > cert.NotAfter.Unix() {
		ctx.err(400, fmt.Errorf("Given certificate is expired or not yet valid"))
		// FIXME: Maybe blow away the url => cert data cache at this point to
		// avoid having to restart the server in the case where the cert has
		// been updated after expiry? Probably in some other failure cases, too.
		return false
	}

	validSAN := false

	for _, san := range cert.Subject.Names {
		if san.Value == "echo-api.amazon.com" {
			validSAN = true
			break
		}
	}

	if !validSAN {
		ctx.err(400, fmt.Errorf("No valid Subject Alternate Names included in given certificate"))
		return false
	}

	// FIXME: At this point we should verify that all certs in the chain combine
	// to create a chain of trust to a trusted root CA.

	publicKey := cert.PublicKey
	encodedSignature := ctx.req.Header.Get("Signature")
	decodedSignature, err := base64.StdEncoding.DecodeString(encodedSignature)
	if err != nil {
		ctx.err(400, fmt.Errorf("Couldn't decode base64 Signature header (%v)", encodedSignature))
		return false
	}

	var buf bytes.Buffer
	hash := sha1.New()
	_, err = io.Copy(hash, io.TeeReader(ctx.req.Body, &buf))
	if err != nil {
		ctx.err(500, fmt.Errorf("IO error: %v", err))
		return false
	}

	ctx.req.Body = ioutil.NopCloser(&buf)

	err = rsa.VerifyPKCS1v15(publicKey.(*rsa.PublicKey), crypto.SHA1, hash.Sum(nil), decodedSignature)
	if err != nil {
		ctx.err(400, fmt.Errorf("Signature verification failed: %v", err))
		return false
	}

	return true
}

func (ctx *requestContext) validateCertChainURL(urlString string) bool {
	url, err := url.Parse(urlString)
	if err != nil {
		ctx.err(400, fmt.Errorf("Invalid SignatureCertChainUrl (%v)", urlString))
		return false
	}

	// FIXME: This should technically be a case-insensitive match.
	if url.Scheme != "https" {
		ctx.err(400, fmt.Errorf("Invalid SignatureCertChainUrl scheme (%v)", urlString))
		return false
	}

	// FIXME: This should technically be a case-insensitive match.
	if url.Host != "s3.amazonaws.com" && url.Host != "s3.amazonaws.com:443" {
		ctx.err(400, fmt.Errorf("Invalid SignatureCertChainUrl hostname (%v)", urlString))
		return false
	}

	// This, on the other hand, *is* supposed to be a case-sensitive match.
	if !strings.HasPrefix(url.Path, "/echo.api/") {
		ctx.err(400, fmt.Errorf("Invalid SignatureCertChainUrl path (%v)", urlString))
		return false
	}

	return true
}

func (ctx *requestContext) validateApplicationID(req *ask.Request) bool {
	if req.Session.Application.ApplicationID != ctx.skill.ApplicationID {
		ctx.err(400, fmt.Errorf("Expected application ID %v, got %v",
			ctx.skill.ApplicationID, req.Session.Application.ApplicationID))
		return false
	}
	return true
}

func (ctx *requestContext) validateTimestamp(req *ask.Request) bool {
	ts, err := time.Parse(time.RFC3339Nano, req.Body.Timestamp)
	if err != nil {
		ctx.err(400, fmt.Errorf("Bad request timestamp %v", req.Body.Timestamp))
		return false
	}

	if time.Since(ts) > time.Duration(30)*time.Second {
		ctx.err(400, fmt.Errorf("Timestamp is too old"))
		return false
	}

	return true
}

func (ctx *requestContext) validateRequestType(req *ask.Request) bool {
	if !req.IsLaunchRequest() && !req.IsIntentRequest() && !req.IsSessionEndedRequest() {
		ctx.err(400, fmt.Errorf("Request type '%v' invalid", req.Body.Type))
		return false
	}
	return true
}

func readCertAtURL(urlString string) ([]byte, error) {
	if bytes, ok := certURLStringToBytesMap[urlString]; ok {
		return bytes, nil
	}

	resp, err := http.Get(urlString)
	if err != nil {
		return nil, fmt.Errorf("Couldn't fetch Amazon cert file: %v", err)
	}

	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Couldn't read Amazon cert file: %v", err)
	}

	certURLStringToBytesMap[urlString] = bytes

	return bytes, nil
}
