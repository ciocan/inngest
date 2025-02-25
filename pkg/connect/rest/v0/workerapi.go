package v0

import (
	"github.com/gogo/protobuf/proto"
	"github.com/inngest/inngest/pkg/connect/auth"
	"github.com/inngest/inngest/pkg/publicerr"
	"github.com/inngest/inngest/proto/gen/connect/v1"
	"io"
	"net/http"
)

func (a *connectApiRouter) start(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	hashedSigningKey := r.Header.Get("Authorization")
	{
		if hashedSigningKey == "" && !a.Dev {
			_ = publicerr.WriteHTTP(w, publicerr.Errorf(401, "missing Authorization header"))
			return
		}

		// Remove "Bearer " prefix
		hashedSigningKey = hashedSigningKey[7:]
	}

	envOverride := r.Header.Get("X-Inngest-Env")

	res, err := a.RequestAuther.AuthenticateRequest(ctx, hashedSigningKey, envOverride)
	if err != nil {
		_ = publicerr.WriteHTTP(w, publicerr.Wrap(err, 401, "authentication failed"))
		return
	}

	if res == nil {
		_ = publicerr.WriteHTTP(w, publicerr.Errorf(401, "authentication failed"))
		return
	}

	byt, err := io.ReadAll(io.LimitReader(r.Body, 1024*1024))
	if err != nil {
		_ = publicerr.WriteHTTP(w, publicerr.Wrap(err, 400, "could not read request body"))
		return
	}

	reqBody := &connect.StartRequest{}
	if len(byt) > 0 {
		if err := proto.Unmarshal(byt, reqBody); err != nil {
			_ = publicerr.WriteHTTP(w, publicerr.Wrap(err, 400, "could not unmarshal request"))
			return
		}
	}

	token, err := a.Signer.SignSessionToken(res.AccountID, res.EnvID, auth.DefaultExpiry)
	if err != nil {
		_ = publicerr.WriteHTTP(w, publicerr.Wrap(err, 500, "could not sign session token"))
		return
	}

	gatewayGroup, gatewayUrl, err := a.ConnectGatewayRetriever.RetrieveGateway(ctx, res.AccountID, res.EnvID, reqBody.ExcludeGateways)
	if err != nil {
		_ = publicerr.WriteHTTP(w, publicerr.Wrap(err, 500, "could not retrieve gateway"))
		return
	}

	msg, err := proto.Marshal(&connect.StartResponse{
		GatewayEndpoint: gatewayUrl.String(),
		GatewayGroup:    gatewayGroup,
		SessionToken:    token,
		SyncToken:       hashedSigningKey,
	})
	if err != nil {
		_ = publicerr.WriteHTTP(w, publicerr.Wrap(err, 500, "could not marshal response"))
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(msg)
}

func (a *connectApiRouter) flushBuffer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	hashedSigningKey := r.Header.Get("Authorization")
	{
		if hashedSigningKey == "" && !a.Dev {
			_ = publicerr.WriteHTTP(w, publicerr.Errorf(401, "missing Authorization header"))
			return
		}

		// Remove "Bearer " prefix
		hashedSigningKey = hashedSigningKey[7:]
	}

	envOverride := r.Header.Get("X-Inngest-Env")

	res, err := a.RequestAuther.AuthenticateRequest(ctx, hashedSigningKey, envOverride)
	if err != nil {
		_ = publicerr.WriteHTTP(w, publicerr.Wrap(err, 401, "authentication failed"))
		return
	}

	if res == nil {
		_ = publicerr.WriteHTTP(w, publicerr.Errorf(401, "authentication failed"))
		return
	}

	byt, err := io.ReadAll(io.LimitReader(r.Body, 1024*1024))
	if err != nil {
		_ = publicerr.WriteHTTP(w, publicerr.Wrap(err, 400, "could not read request body"))
		return
	}

	reqBody := &connect.SDKResponse{}
	if len(byt) > 0 {
		if err := proto.Unmarshal(byt, reqBody); err != nil {
			_ = publicerr.WriteHTTP(w, publicerr.Wrap(err, 400, "could not unmarshal request"))
			return
		}
	}

	// Marshal response before notifying executor, marshaling should never fail
	msg, err := proto.Marshal(&connect.FlushResponse{
		RequestId: reqBody.RequestId,
	})
	if err != nil {
		_ = publicerr.WriteHTTP(w, publicerr.Wrap(err, 500, "could not marshal response"))
		return
	}

	if err := a.ConnectResponseNotifier.NotifyExecutor(ctx, reqBody); err != nil {
		_ = publicerr.WriteHTTP(w, publicerr.Wrap(err, 500, "could not notify executor"))
		return
	}

	// Send response once executor was notified
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(msg)
}
