package fw4

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel/trace"

	"core/internal/logger"
)

func parseIntToken(token string) (int32, error) {
	result, err := strconv.ParseInt(token, 10, 32)
	return int32(result), err
}

func parseUIntToken(token string) (uint32, error) {
	result, err := parseIntToken(token)
	return uint32(result), err
}

func parseHexBlobIntImpl(token string, requireZeroPad bool) (int32, error) {
	if requireZeroPad && len(token)%2 != 0 {
		return 0, errors.New("invalid hex blob value: odd length")
	}
	result, err := strconv.ParseUint(token, 16, 32)
	return int32(result), err
}

func parseHexBlobInt(token string) (int32, error) {
	return parseHexBlobIntImpl(token, true)
}

func parseHexIntToken(token string) (int32, error) {
	return parseHexBlobIntImpl(token, false)
}

func parseHexUIntToken(token string) (uint32, error) {
	result, err := parseHexBlobIntImpl(token, false)
	return uint32(result), err
}

func checkFw4Prefix(value string, fw4prefix string) ([]string, error) {
	// The required token count is always 7 by chance, but the tokens are not always the same
	const requiredTokenCount = 7

	if !strings.HasPrefix(value, fw4prefix) {
		return nil, fmt.Errorf("expected FW4 tag but input does not start with %q", fw4prefix)
	}

	// Split into up to one more than required tokens. The last token is the optional (unsplit) extension part.
	tokens := strings.SplitN(value[len(fw4prefix):], ";", requiredTokenCount+1)
	if len(tokens) < requiredTokenCount {
		return tokens, errors.New("FW4 tag does not contain enough semicolon-separated components")
	}
	return tokens, nil
}

func ParseXDynatrace(header string) (Fw4Tag, error) {
	result := EmptyTag()

	tokens, err := checkFw4Prefix(header, "FW4;")
	if err != nil {
		return result, err
	}

	result.ClusterID, err = parseIntToken(tokens[0])
	if err != nil {
		return result, fmt.Errorf("invalid value for ClusterId: %w", err)
	}

	result.ServerID, err = parseIntToken(tokens[1])
	if err != nil {
		return result, fmt.Errorf("invalid value for ServerId: %w", err)
	}

	result.agentID, err = parseIntToken(tokens[2])
	if err != nil {
		return result, fmt.Errorf("invalid value for agentId: %w", err)
	}

	result.tagID, err = parseIntToken(tokens[3])
	if err != nil {
		return result, fmt.Errorf("invalid value for tagId: %w", err)
	}

	result.encodedLinkID, err = parseUIntToken(tokens[4])
	if err != nil {
		return result, fmt.Errorf("invalid value for encodedLinkId: %w", err)
	}

	result.TenantID, err = parseIntToken(tokens[5])
	if err != nil {
		return result, fmt.Errorf("invalid value for TenantId: %w", err)
	}

	result.PathInfo, err = parseUIntToken(tokens[6])
	if err != nil {
		return result, fmt.Errorf("invalid value for PathInfo: %w", err)
	}

	if len(tokens) > 7 {
		err = parseExtensions(&result, tokens[7])
	}
	return result, err
}

func ParseTracestateEntryValue(entryValue string) (Fw4Tag, error) {
	result := EmptyTag()

	tokens, err := checkFw4Prefix(entryValue, "fw4;")
	if err != nil {
		return result, err
	}

	result.ServerID, err = parseHexIntToken(tokens[0])
	if err != nil {
		return result, fmt.Errorf("invalid value for ServerId: %w", err)
	}

	result.agentID, err = parseHexIntToken(tokens[1])
	if err != nil {
		return result, fmt.Errorf("invalid value for agentId: %w", err)
	}

	result.tagID, err = parseHexIntToken(tokens[2])
	if err != nil {
		return result, fmt.Errorf("invalid value for tagId: %w", err)
	}

	result.encodedLinkID, err = parseHexUIntToken(tokens[3])
	if err != nil {
		return result, fmt.Errorf("invalid value for link ID: %w", err)
	}

	linkIdBits, err := parseHexUIntToken(tokens[4])
	if err != nil {
		return result, fmt.Errorf("invalid value for ignored bit: %w", err)
	}
	result.encodedLinkID |= linkIdBits << 31

	linkIdBits, err = parseHexUIntToken(tokens[5])
	if err != nil {
		return result, fmt.Errorf("invalid value for sampling exponent: %w", err)
	}
	// TODO: We could override the sampled bit here if the exponent is too large
	result.encodedLinkID |= linkIdBits << 27

	result.PathInfo, err = parseHexUIntToken(tokens[6])
	if err != nil {
		return result, fmt.Errorf("invalid value for PathInfo: %w", err)
	}

	if len(tokens) > 7 {
		err = parseExtensions(&result, tokens[7])
	}
	return result, err
}

func parseExtensions(result *Fw4Tag, extensionPart string) error {
	// This code allows a trailing checksum with empty content
	const checksumLen = 4
	if len(extensionPart) < checksumLen || len(extensionPart) > checksumLen && extensionPart[checksumLen] != ';' {
		return fmt.Errorf("extension checksum must be exactly %v ASCII bytes long", checksumLen)
	}
	checksum, err := strconv.ParseUint(extensionPart[:checksumLen], 16, 16)
	if err != nil {
		return fmt.Errorf("unparsable extension checksum: %w", err)
	}
	// Hash the part after the checksum, including the trailing ";" (if any)
	actualChecksum := uint16(latin1MurmurHash2_64A(extensionPart[checksumLen:]))
	if uint16(checksum) != actualChecksum {
		return fmt.Errorf("checksum mismatch: expected=%x actual=%x", checksum, actualChecksum)
	}

	for _, token := range strings.Split(extensionPart[checksumLen+1:], ";") {
		posH := strings.IndexByte(token, 'h')
		if posH <= 0 {
			// Note: libcorrelation recognizes a few types other than "h" internally, but they are never used,
			// so we don't implement them at all.
			continue
		}
		extID, err := strconv.ParseUint(token[:posH], 10, 32)
		if err != nil {
			return fmt.Errorf("unparsable extension ID: %w", err)
		}
		err = parseAndSetExtValue(result, tlvId(extID), token[posH+1:])
		if err != nil {
			return fmt.Errorf("unparsable extension value for ID %v: %w", extID, err)
		}
	}

	return nil
}

var log = logger.NewComponentLogger("fw4")

func parseAndSetExtValue(result *Fw4Tag, tlvId tlvId, hexVal string) error {
	var err error

	switch tlvId {
	case tlvIdCustomBlob:
		if len(hexVal) > hex.EncodedLen(maxBlobLen) {
			return errors.New("invalid hex blob value: too long")
		}
		decoded, err := hex.DecodeString(hexVal)
		if err != nil {
			return err
		}
		result.CustomBlob = string(decoded)
	case tlvIdTagDepth:
		result.tagDepth, err = parseHexBlobInt(hexVal)
	case tlvIdEntryAgentId:
		result.entryAgentID, err = parseHexBlobInt(hexVal)
	case tlvIdEntryTagId:
		result.entryTagID, err = parseHexBlobInt(hexVal)
	case tlvIdPayloadBitset:
		result.payloadBitset, err = parseHexBlobInt(hexVal)
	case tlvIdTraceId:
		result.TraceID, err = trace.TraceIDFromHex(hexVal)
	case tlvIdSpanId:
		result.SpanID, err = trace.SpanIDFromHex(hexVal)
	default:
		// Unknown extension values are ignored without error
		log.Debugf("Ignoring unknown extension value: %v", tlvId)
	}

	return err
}

func (fw4 Fw4Tag) ToXDynatrace() string {
	var sb strings.Builder

	fmt.Fprintf(
		&sb,
		"FW4;%d;%d;%d;%d;%d;%d;%d",
		fw4.ClusterID,
		fw4.ServerID,
		fw4.agentID,
		fw4.tagID,
		int32(fw4.encodedLinkID),
		fw4.TenantID,
		int32(fw4.PathInfo))
	encodeExtensions(fw4, &sb)

	return sb.String()
}

func (fw4 Fw4Tag) ToTracestateEntryValue() string {
	var sb strings.Builder

	ignoredStr := '0'
	if fw4.IsIgnored() {
		ignoredStr = '1'
	}

	fmt.Fprintf(
		&sb,
		"fw4;%x;%x;%x;%x;%c;%x;%x",
		uint32(fw4.ServerID),
		uint32(fw4.agentID),
		uint32(fw4.tagID),
		uint32(fw4.linkID()),
		ignoredStr,
		uint32(fw4.samplingExponent()),
		fw4.PathInfo)
	encodeExtensions(fw4, &sb)

	return sb.String()
}

//nolint:errcheck
func encodeExtensions(fw4 Fw4Tag, fw4Sb *strings.Builder) {
	hasExt := false
	var extSb strings.Builder
	beginExtField := func(extId tlvId) {
		if !hasExt {
			hasExt = true
		}
		extSb.WriteByte(';')
		extSb.WriteString(strconv.Itoa(int(extId)))
		extSb.WriteByte('h')
	}

	writeHexIntField := func(extId tlvId, v int32) {
		beginExtField(extId)
		h := strconv.FormatUint(uint64(uint32(v)), 16)
		if len(h)%2 != 0 {
			extSb.WriteByte('0')
		}
		extSb.WriteString(h)
	}

	if len(fw4.CustomBlob) > 0 {
		beginExtField(tlvIdCustomBlob)
		hex.NewEncoder(&extSb).Write([]byte(fw4.CustomBlob))
	}

	if fw4.HasTagDepth() {
		writeHexIntField(tlvIdTagDepth, fw4.tagDepth)
	}

	if fw4.HasEntryAgentID() {
		writeHexIntField(tlvIdEntryAgentId, fw4.entryAgentID)
	}

	if fw4.HasEntryTagID() {
		writeHexIntField(tlvIdEntryTagId, fw4.entryTagID)
	}

	if fw4.payloadBitset != 0 {
		writeHexIntField(tlvIdPayloadBitset, fw4.payloadBitset)
	}

	if fw4.TraceID.IsValid() {
		beginExtField(tlvIdTraceId)
		hex.NewEncoder(&extSb).Write(fw4.TraceID[:])
	}

	if fw4.SpanID.IsValid() {
		beginExtField(tlvIdSpanId)
		hex.NewEncoder(&extSb).Write(fw4.SpanID[:])
	}

	if hasExt {
		extStr := extSb.String()
		checksum := uint16(latin1MurmurHash2_64A(extStr))
		fmt.Fprintf(fw4Sb, ";%04x%s", checksum, extStr)
	}
}

func (fw4 Fw4Tag) TraceStateKey() string {
	return fmt.Sprintf("%x-%x@dt", fw4.TenantID, fw4.ClusterID)
}
