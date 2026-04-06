package steam

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"sync/atomic"
	"time"

	. "github.com/paralin/go-steam/protocol"
	. "github.com/paralin/go-steam/protocol/protobuf"
	. "github.com/paralin/go-steam/protocol/steamlang"
	"github.com/paralin/go-steam/steamid"
	"github.com/golang/protobuf/proto"
)

type Auth struct {
	client  *Client
	Details *LogOnDetails

	authSession   *CAuthentication_BeginAuthSessionViaCredentials_Response
	Authenticator Authenticator
}

// Authenticator provides custom Steam Guard code input for credential-based auth.
type Authenticator interface {
	GetCode(EAuthSessionGuardType, func(string, EAuthSessionGuardType) error) error
}

type SentryHash []byte

type LogOnDetails struct {
	Username string

	// If logging into an account without a login key, the account's password.
	Password string

	// If you have a Steam Guard email code, you can provide it here.
	AuthCode string

	// If you have a Steam Guard mobile two-factor authentication code, you can provide it here.
	TwoFactorCode  string
	SentryFileHash SentryHash
	LoginKey       string

	// true if you want to get a login key which can be used in lieu of
	// a password for subsequent logins. false or omitted otherwise.
	ShouldRememberPassword bool

	// AccessToken from a previous credential-based auth session.
	AccessToken string
	// RefreshToken from a previous credential-based auth session.
	// When set, LogOn() uses this instead of Password.
	RefreshToken string
	// GuardData from a previous credential-based auth session.
	GuardData string
}

// Log on with the given details. You must always specify username and
// password OR username and loginkey. For the first login, don't set an authcode or a hash and you'll
//  receive an error (EResult_AccountLogonDenied)
// and Steam will send you an authcode. Then you have to login again, this time with the authcode.
// Shortly after logging in, you'll receive a MachineAuthUpdateEvent with a hash which allows
// you to login without using an authcode in the future.
//
// If you don't use Steam Guard, username and password are enough.
//
// After the event EMsg_ClientNewLoginKey is received you can use the LoginKey
// to login instead of using the password.
func (a *Auth) LogOn(details *LogOnDetails) {
	if details.Username == "" {
		panic("Username must be set!")
	}
	if details.Password == "" && details.LoginKey == "" && details.RefreshToken == "" {
		panic("Password, LoginKey or RefreshToken must be set!")
	}

	logon := new(CMsgClientLogon)
	logon.AccountName = &details.Username
	logon.Password = &details.Password
	if details.RefreshToken != "" {
		logon.AccessToken = proto.String(details.RefreshToken)
	}
	if details.AuthCode != "" {
		logon.AuthCode = proto.String(details.AuthCode)
	}
	if details.TwoFactorCode != "" {
		logon.TwoFactorCode = proto.String(details.TwoFactorCode)
	}
	logon.ClientLanguage = proto.String("english")
	logon.ProtocolVersion = proto.Uint32(MsgClientLogon_CurrentProtocol)
	logon.ShaSentryfile = details.SentryFileHash
	if details.LoginKey != "" {
		logon.LoginKey = proto.String(details.LoginKey)
	}
	if details.ShouldRememberPassword {
		logon.ShouldRememberPassword = proto.Bool(details.ShouldRememberPassword)
	}

	atomic.StoreUint64(&a.client.steamId, steamid.NewIdAdv(0, 1, int32(EUniverse_Public), EAccountType_Individual).ToUint64())

	a.client.Write(NewClientMsgProtobuf(EMsg_ClientLogon, logon))
}

func (a *Auth) HandlePacket(packet *Packet) {
	switch packet.EMsg {
	case EMsg_ClientLogOnResponse:
		a.handleLogOnResponse(packet)
	case EMsg_ClientNewLoginKey:
		a.handleLoginKey(packet)
	case EMsg_ClientSessionToken:
	case EMsg_ClientLoggedOff:
		a.handleLoggedOff(packet)
	case EMsg_ClientUpdateMachineAuth:
		a.handleUpdateMachineAuth(packet)
	case EMsg_ClientAccountInfo:
		a.handleAccountInfo(packet)
	case EMsg_ClientWalletInfoUpdate:
	case EMsg_ClientRequestWebAPIAuthenticateUserNonceResponse:
	case EMsg_ClientMarketingMessageUpdate:
	}
}

func (a *Auth) handleLogOnResponse(packet *Packet) {
	if !packet.IsProto {
		a.client.Fatalf("Got non-proto logon response!")
		return
	}

	body := new(CMsgClientLogonResponse)
	msg := packet.ReadProtoMsg(body)

	result := EResult(body.GetEresult())
	if result == EResult_OK {
		atomic.StoreInt32(&a.client.sessionId, msg.Header.Proto.GetClientSessionid())
		atomic.StoreUint64(&a.client.steamId, msg.Header.Proto.GetSteamid())
		if body.WebapiAuthenticateUserNonce != nil {
			a.client.Web.webLoginKey = *body.WebapiAuthenticateUserNonce
		}

		go a.client.heartbeatLoop(time.Duration(body.GetOutOfGameHeartbeatSeconds()))

		a.client.Emit(&LoggedOnEvent{
			Result:         EResult(body.GetEresult()),
			ExtendedResult: EResult(body.GetEresultExtended()),
			AccountFlags:   EAccountFlags(body.GetAccountFlags()),
			ClientSteamId:  steamid.SteamId(body.GetClientSuppliedSteamid()),
			Body:           body,
		})
	} else if result == EResult_Fail || result == EResult_ServiceUnavailable || result == EResult_TryAnotherCM {
		// some error on Steam's side, we'll get an EOF later
		a.client.Emit(&SteamFailureEvent{
			Result: EResult(body.GetEresult()),
		})
	} else {
		a.client.Emit(&LogOnFailedEvent{
			Result: EResult(body.GetEresult()),
		})
		a.client.Disconnect()
	}
}

func (a *Auth) handleLoginKey(packet *Packet) {
	body := new(CMsgClientNewLoginKey)
	packet.ReadProtoMsg(body)
	a.client.Write(NewClientMsgProtobuf(EMsg_ClientNewLoginKeyAccepted, &CMsgClientNewLoginKeyAccepted{
		UniqueId: proto.Uint32(body.GetUniqueId()),
	}))
	a.client.Emit(&LoginKeyEvent{
		UniqueId: body.GetUniqueId(),
		LoginKey: body.GetLoginKey(),
	})
}

func (a *Auth) handleLoggedOff(packet *Packet) {
	result := EResult_Invalid
	if packet.IsProto {
		body := new(CMsgClientLoggedOff)
		packet.ReadProtoMsg(body)
		result = EResult(body.GetEresult())
	} else {
		body := new(MsgClientLoggedOff)
		packet.ReadClientMsg(body)
		result = body.Result
	}
	a.client.Emit(&LoggedOffEvent{Result: result})
}

func (a *Auth) handleUpdateMachineAuth(packet *Packet) {
	body := new(CMsgClientUpdateMachineAuth)
	packet.ReadProtoMsg(body)
	hash := sha1.New()
	hash.Write(packet.Data)
	sha := hash.Sum(nil)

	msg := NewClientMsgProtobuf(EMsg_ClientUpdateMachineAuthResponse, &CMsgClientUpdateMachineAuthResponse{
		ShaFile: sha,
	})
	msg.SetTargetJobId(packet.SourceJobId)
	a.client.Write(msg)

	a.client.Emit(&MachineAuthUpdateEvent{sha})
}

func (a *Auth) handleAccountInfo(packet *Packet) {
	body := new(CMsgClientAccountInfo)
	packet.ReadProtoMsg(body)
	a.client.Emit(&AccountInfoEvent{
		PersonaName:          body.GetPersonaName(),
		Country:              body.GetIpCountry(),
		CountAuthedComputers: body.GetCountAuthedComputers(),
		AccountFlags:         EAccountFlags(body.GetAccountFlags()),
		FacebookId:           body.GetFacebookId(),
		FacebookName:         body.GetFacebookName(),
	})
}

// LogOnCredentials performs the modern Steam credential-based authentication flow.
// It sends ClientHello, fetches the RSA public key, encrypts the password,
// begins an auth session, handles Steam Guard, polls for tokens, and finally
// calls LogOn with the obtained refresh token.
func (a *Auth) LogOnCredentials(details *LogOnDetails) {
	if details.Username == "" {
		panic("Username must be set!")
	}
	if details.Password == "" && details.LoginKey == "" {
		panic("Password or LoginKey must be set!")
	}

	atomic.StoreUint64(&a.client.steamId, steamid.NewIdAdv(0, 1, int32(EUniverse_Public), EAccountType_Individual).ToUint64())

	hello := &CMsgClientHello{ProtocolVersion: proto.Uint32(MsgClientLogon_CurrentProtocol)}
	a.client.Write(NewClientMsgProtobuf(EMsg_ClientHello, hello))

	a.Details = details
	a.getRSAKey(details.Username)
}

func encryptPassword(pwd string, key *CAuthentication_GetPasswordRSAPublicKey_Response) (string, error) {
	var n big.Int
	n.SetString(*key.PublickeyMod, 16)

	exp, err := strconv.ParseInt(*key.PublickeyExp, 16, 32)
	if err != nil {
		return "", err
	}

	pub := rsa.PublicKey{N: &n, E: int(exp)}
	rsaOut, err := rsa.EncryptPKCS1v15(rand.Reader, &pub, []byte(pwd))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(rsaOut), nil
}

func (a *Auth) getRSAKey(accountName string) {
	req := new(CAuthentication_GetPasswordRSAPublicKey_Request)
	req.AccountName = &accountName

	msg := NewClientMsgProtobuf(EMsg_ServiceMethodCallFromClientNonAuthed, req)
	jobname := "Authentication.GetPasswordRSAPublicKey#1"
	msg.Header.Proto.TargetJobName = &jobname
	jobID := a.client.GetNextJobId()
	msg.SetSourceJobId(jobID)

	a.client.JobMutex.Lock()
	a.client.JobHandlers[uint64(jobID)] = a.beginAuthSession
	a.client.JobMutex.Unlock()

	a.client.Write(msg)
}

func (a *Auth) beginAuthSession(packet *Packet) error {
	body := new(CAuthentication_GetPasswordRSAPublicKey_Response)
	_ = packet.ReadProtoMsg(body)

	crypt, err := encryptPassword(a.Details.Password, body)
	if err != nil {
		return err
	}

	deviceFriendlyName := "DESKTOP-HELLO"
	platformType := EAuthTokenPlatformType_k_EAuthTokenPlatformType_SteamClient.Enum()

	deviceDetails := CAuthentication_DeviceDetails{
		DeviceFriendlyName: &deviceFriendlyName,
		PlatformType:       platformType,
		OsType:             proto.Int32(16),
	}

	req := CAuthentication_BeginAuthSessionViaCredentials_Request{
		AccountName:         &a.Details.Username,
		EncryptedPassword:   &crypt,
		EncryptionTimestamp: body.Timestamp,
		Persistence:         ESessionPersistence_k_ESessionPersistence_Persistent.Enum(),
		WebsiteId:           proto.String("Client"),
		DeviceDetails:       &deviceDetails,
	}

	// Pass guard data to skip email code if machine is already trusted
	if a.Details.GuardData != "" {
		req.GuardData = &a.Details.GuardData
	}

	msg := NewClientMsgProtobuf(EMsg_ServiceMethodCallFromClientNonAuthed, &req)
	jobname := "Authentication.BeginAuthSessionViaCredentials#1"
	msg.Header.Proto.TargetJobName = &jobname
	jobID := a.client.GetNextJobId()
	msg.SetSourceJobId(jobID)

	a.client.JobMutex.Lock()
	a.client.JobHandlers[uint64(jobID)] = a.handleAuthSession
	a.client.JobMutex.Unlock()

	a.client.Write(msg)
	return nil
}

func (a *Auth) handleAuthSession(packet *Packet) error {
	body := new(CAuthentication_BeginAuthSessionViaCredentials_Response)
	_ = packet.ReadProtoMsg(body)

	a.authSession = body

	var codeType EAuthSessionGuardType
	for _, confirmation := range body.AllowedConfirmations {
		switch *confirmation.ConfirmationType {
		case EAuthSessionGuardType_k_EAuthSessionGuardType_None,
			EAuthSessionGuardType_k_EAuthSessionGuardType_MachineToken:
			return a.pollAuthSession()
		case EAuthSessionGuardType_k_EAuthSessionGuardType_EmailCode:
			codeType = EAuthSessionGuardType_k_EAuthSessionGuardType_EmailCode
			fallthrough
		case EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceCode:
			if codeType == 0 {
				codeType = EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceCode
			}

			if a.Authenticator != nil {
				return a.Authenticator.GetCode(codeType, a.updateAuthSession)
			}

			go func() {
				var code string
				fmt.Println("Enter Code:")
				_, _ = fmt.Scanln(&code)
				a.updateAuthSession(code, codeType)
			}()

		case EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceConfirmation:
		case EAuthSessionGuardType_k_EAuthSessionGuardType_EmailConfirmation:
		case EAuthSessionGuardType_k_EAuthSessionGuardType_LegacyMachineAuth:
		case EAuthSessionGuardType_k_EAuthSessionGuardType_Unknown:
		}
	}

	return nil
}

func (a *Auth) updateAuthSession(code string, codeType EAuthSessionGuardType) error {
	req := CAuthentication_UpdateAuthSessionWithSteamGuardCode_Request{
		ClientId: a.authSession.ClientId,
		Steamid:  a.authSession.Steamid,
		Code:     &code,
		CodeType: &codeType,
	}

	msg := NewClientMsgProtobuf(EMsg_ServiceMethodCallFromClientNonAuthed, &req)
	jobname := "Authentication.UpdateAuthSessionWithSteamGuardCode#1"
	msg.Header.Proto.TargetJobName = &jobname
	jobID := a.client.GetNextJobId()
	msg.SetSourceJobId(jobID)

	a.client.JobMutex.Lock()
	a.client.JobHandlers[uint64(jobID)] = a.handleAuthSessionUpdate
	a.client.JobMutex.Unlock()

	a.client.Write(msg)
	return nil
}

func (a *Auth) handleAuthSessionUpdate(packet *Packet) error {
	body := new(CAuthentication_UpdateAuthSessionWithSteamGuardCode_Response)
	_ = packet.ReadProtoMsg(body)

	return a.pollAuthSession()
}

func (a *Auth) pollAuthSession() error {
	req := CAuthentication_PollAuthSessionStatus_Request{
		ClientId:  a.authSession.ClientId,
		RequestId: a.authSession.RequestId,
	}

	msg := NewClientMsgProtobuf(EMsg_ServiceMethodCallFromClientNonAuthed, &req)
	jobname := "Authentication.PollAuthSessionStatus#1"
	msg.Header.Proto.TargetJobName = &jobname
	jobID := a.client.GetNextJobId()
	msg.SetSourceJobId(jobID)

	a.client.JobMutex.Lock()
	a.client.JobHandlers[uint64(jobID)] = a.handlePollResponse
	a.client.JobMutex.Unlock()

	a.client.Write(msg)
	return nil
}

func (a *Auth) handlePollResponse(packet *Packet) error {
	body := new(CAuthentication_PollAuthSessionStatus_Response)
	_ = packet.ReadProtoMsg(body)

	if body.RefreshToken == nil {
		return errors.New("AuthSession PollError")
	}

	a.Details.AccessToken = *body.AccessToken
	a.Details.RefreshToken = *body.RefreshToken
	if body.NewGuardData != nil {
		a.Details.GuardData = *body.NewGuardData
	}

	a.LogOn(a.Details)
	return nil
}
