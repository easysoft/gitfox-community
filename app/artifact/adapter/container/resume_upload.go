// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package container

import (
	"bytes"
	"context"
	"encoding"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/easysoft/gitfox-community/app/api/utils/req"
	"github.com/easysoft/gitfox-community/app/artifact/adapter"
	model "github.com/easysoft/gitfox-community/app/model/artifact"

	artifactStore "github.com/easysoft/gitfox-community/app/store/artifact"
	pkgerrors "github.com/easysoft/gitfox-community/pkg/errors"

	"github.com/easysoft/gitfox-community/pkg/logger"
	"github.com/google/uuid"
	"github.com/opencontainers/go-digest"

	_ "crypto/sha256"
)

const (
	errInvalidToken = "incorrect state"
)

type ResumeRequest struct {
	artStore artifactStore.Interface
	view     *adapter.ViewDescriptor

	digester digest.Digester

	SessionID string
	RepoName  string
	Ref       string

	State *blobUploadState
}

type blobUploadState struct {
	// name is the primary repository under which the blob will be linked.
	Name string

	// UUID identifies the upload.
	SessionID string

	// offset contains the current progress of the upload.
	Offset int64

	// StartedAt is the original start time of the upload.
	StartedAt int64

	// Ref is the storage blob name
	Ref string
}

func (bs *blobUploadState) Pack(key string) (string, error) {
	return hmacKey(key).packUploadState(*bs)
}

type progressStatus string

const (
	progressNotStarted progressStatus = "not_started"
	progressInComplete progressStatus = "incomplete"
	progressCompleted  progressStatus = "completed"
)

type metadataProgress struct {
	Digester string         `json:"digester"`
	Status   progressStatus `json:"status"`
}

// InitResumeRequest generate a random session_id for upload,
// init an empty resume request
func InitResumeRequest(repoName string, artStore artifactStore.Interface, view *adapter.ViewDescriptor) *ResumeRequest {
	newSessionID := uuid.New()
	newRef := strings.ReplaceAll(HashUUID(newSessionID, repoName), "-", "")

	bs := blobUploadState{
		Name:      repoName,
		SessionID: newSessionID.String(),
		Offset:    0,
		StartedAt: time.Now().UnixMilli(),
		Ref:       newRef,
	}

	return &ResumeRequest{
		view:     view,
		artStore: artStore,

		RepoName:  repoName,
		SessionID: bs.SessionID,
		Ref:       bs.Ref,

		State: &bs,
	}
}

func ParseResumeRequest(ctx context.Context, sessionID, fullRepoName, token string, artStore artifactStore.Interface, view *adapter.ViewDescriptor) (*ResumeRequest, error) {
	req := &ResumeRequest{
		artStore: artStore,
		view:     view,
		RepoName: fullRepoName,
	}

	sid, err := uuid.Parse(sessionID)
	if err != nil {
		logger.GetLogger(ctx).Errorf("invalid session id: %v", err)
		return nil, ErrBlobUploadInvalid.WithDetail(err.Error())
	}
	req.SessionID = sessionID
	req.Ref = strings.ReplaceAll(HashUUID(sid, fullRepoName), "-", "")

	state, err := hmacKey("").unpackUploadState(token)
	if err != nil {
		return nil, ErrBlobUploadInvalid.WithDetail(err.Error())
	}

	if state.SessionID != req.SessionID {
		logger.GetLogger(ctx).Infof("mismatched uuid in upload state: %q != %q", state.SessionID, req.SessionID)
		return nil, ErrBlobUploadInvalid.WithDetail(errInvalidToken)
	}

	if state.Ref != req.Ref {
		logger.GetLogger(ctx).Infof("mismatched ref in upload state: %q != %q", state.Ref, req.Ref)
		return nil, ErrBlobUploadInvalid.WithDetail(errInvalidToken)
	}

	if state.Name != req.RepoName {
		logger.GetLogger(ctx).Infof("mismatched repository name in upload state: %q != %q", state.Name, req.RepoName)
		return nil, ErrBlobUploadInvalid.WithDetail(errInvalidToken)
	}

	status, digester, err := parseProgress(ctx, artStore, view, req.Ref)
	if err != nil {
		return nil, err
	}

	if status == progressCompleted {
		return nil, ErrBlobUploadInvalid.WithDetail("upload is already completed")
	}

	req.digester = digester
	req.State = &state
	return req, nil
}

func parseProgress(ctx context.Context, artStore artifactStore.Interface, view *adapter.ViewDescriptor, ref string) (progressStatus, digest.Digester, error) {
	var digester = digest.Canonical.Digester()
	var status progressStatus

	dbBlob, err := artStore.Blobs().GetByRef(ctx, ref, view.StorageID)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrResourceNotFound) {
			status = progressNotStarted
			return status, digester, nil
		} else {
			return "", nil, ErrBlobUploadInvalid.WithDetail(err.Error())
		}
	}

	if dbBlob.Metadata == "" {
		// this should never have happened
		logger.GetLogger(ctx).Infof("upload metadata not found in record")
		return "", nil, ErrBlobUploadInvalid.WithDetail("upload metadata not found in server")
	}

	var upMeta metadataProgress
	if err = json.NewDecoder(bytes.NewReader([]byte(dbBlob.Metadata))).Decode(&upMeta); err != nil {
		return "", nil, ErrBlobUploadInvalid.WithDetail("parse upload metadata failed")
	}

	digesterSnapshot, err := base64.StdEncoding.DecodeString(upMeta.Digester)
	if err != nil {
		return "", nil, ErrBlobUploadInvalid.WithDetail("resume upload failed")
	}
	rd, ok := digester.Hash().(encoding.BinaryUnmarshaler)
	if !ok {
		return "", nil, ErrBlobUploadInvalid.WithDetail("resumable digest not available")
	}
	if err = rd.UnmarshalBinary(digesterSnapshot); err != nil {
		return "", nil, ErrBlobUploadInvalid.WithDetail("resume upload failed")
	}

	return upMeta.Status, digester, nil
}

func (crr *ResumeRequest) Size(ctx context.Context) (int64, error) {
	fInfo, err := crr.view.Store.Stat(ctx, adapter.BlobPath(crr.Ref))
	if err != nil {
		return 0, err
	}
	return fInfo.Size(), nil
}

func (crr *ResumeRequest) AppendWrite(ctx context.Context, r *http.Request) error {
	log := logger.GetLogger(ctx)
	ct := r.Header.Get("Content-Type")
	if ct != "" && ct != "application/octet-stream" {
		return ErrUnknown.WithDetail("bad Content-Type")
	}

	fw, err := adapter.NewStorageBlobWriter(ctx, crr.view.Store, crr.Ref, true)
	if err != nil {
		log.Errorf("【container push】create blob writer failed: ref=%s err=%v", crr.Ref, err)
		return ErrBlobUploadUnknown.WithDetail(err.Error())
	}
	defer func() {
		if err != nil {
			fw.Writer.Cancel(ctx)
		} else {
			fw.Writer.Commit(ctx)
		}
		fw.Writer.Close()
	}()

	cr := r.Header.Get("Content-Range")
	cl := r.Header.Get("Content-Length")
	if cr != "" && cl != "" {
		start, end, err := parseContentRange(cr)
		if err != nil {
			return ErrUnknown.WithDetail(err.Error())
		}
		if start > end || start != fw.Writer.Size() {
			return ErrCodeRangeInvalid
		}

		clInt, err := strconv.ParseInt(cl, 10, 64)
		if err != nil {
			return ErrUnknown.WithDetail(err.Error())
		}
		if clInt != (end-start)+1 {
			return ErrSizeInvalid
		}
	}

	if err = copyFullPayload(ctx, r, fw.Writer, crr.digester); err != nil {
		log.Errorf("【container push】PATCH copy payload failed: session=%s ref=%s err=%v", crr.SessionID, crr.Ref, err)
		return err
	}

	crr.State.Offset = fw.Writer.Size()
	if _, err = crr.saveProgress(ctx, progressInComplete, crr.State.Offset); err != nil {
		return err
	}
	return nil
}

func (crr *ResumeRequest) Finish(ctx context.Context, r *http.Request, dgst *digest.Digest) error {
	log := logger.GetLogger(ctx)
	fw, err := adapter.NewStorageBlobWriter(ctx, crr.view.Store, crr.Ref, true)
	if err != nil {
		log.Errorf("【container push】create blob writer failed: ref=%s err=%v", crr.Ref, err)
		return ErrBlobUploadUnknown.WithDetail(err.Error())
	}
	defer func() {
		if err != nil {
			fw.Writer.Cancel(ctx)
		} else {
			if e := fw.Writer.Commit(ctx); e != nil {
				log.Errorf("【container push】commit chunk failed: ref=%s err=%v", crr.Ref, e)
			}
		}
		fw.Writer.Close()
	}()

	if err = copyFullPayload(ctx, r, fw.Writer, crr.digester); err != nil {
		log.Errorf("【container push】PUT copy payload failed: session=%s ref=%s err=%v", crr.SessionID, crr.Ref, err)
		return err
	}

	completeDigest := crr.digester.Digest()
	if completeDigest.String() != dgst.String() {
		log.Errorf(
			"【container push】digest mismatch: session=%s size=%d got=%s want=%s",
			crr.SessionID, fw.Writer.Size(), completeDigest, dgst,
		)
		return ErrBlobUploadInvalid.WithDetail("digest mismatch")
	}

	dbBlob, err := crr.saveProgress(ctx, progressCompleted, fw.Writer.Size())
	if err != nil {
		log.Errorf("【container push】save progress failed: session=%s err=%v", crr.SessionID, err)
		return err
	}

	asset := model.CreateArtifactAsset{
		Format:      model.ArtifactContainerFormat,
		Path:        dgst.String(),
		ContentType: "application/octet-stream",
		Rank:        model.AssetRankMain,
		BlobID:      dbBlob.ID,
	}

	art, err := crr.artStore.Assets().GetPath(ctx, asset.Path, model.ArtifactContainerFormat)
	if err != nil {
		if _, err = crr.artStore.Assets().Create(ctx, &asset); err != nil {
			log.Errorf("【container push】create asset failed: digest=%s err=%v", dgst, err)
			return err
		}
	} else {
		art.BlobID = dbBlob.ID
		if err = crr.artStore.Assets().Update(ctx, art); err != nil {
			log.Errorf("【container push】update asset failed: digest=%s err=%v", dgst, err)
			return err
		}
	}

	return nil
}

func (crr *ResumeRequest) Remove(ctx context.Context) error {
	blobPath := adapter.BlobPath(crr.Ref)
	_, err := crr.view.Store.Stat(ctx, blobPath)
	if err == nil {
		_ = crr.view.Store.Delete(ctx, blobPath)
		return nil
	}
	return nil
}

func (crr *ResumeRequest) saveProgress(ctx context.Context, status progressStatus, size int64) (*model.ArtifactBlob, error) {
	bm, _ := crr.digester.Hash().(encoding.BinaryMarshaler)
	state, err := bm.MarshalBinary()
	if err != nil {
		return nil, ErrUnknown.WithDetail(err.Error())
	}
	b64State := base64.StdEncoding.EncodeToString(state)
	upMeta := metadataProgress{
		Digester: b64State,
		Status:   status,
	}

	mutateFn := func(blob *model.ArtifactBlob) error {
		jMeta, e := json.Marshal(&upMeta)
		if e != nil {
			return e
		}
		blob.Metadata = string(jMeta)
		blob.Size = size
		return nil
	}

	var dbBlob *model.ArtifactBlob
	if dbBlob, err = crr.createOrUpdateModelBlob(ctx, mutateFn); err != nil {
		return nil, err
	}

	return dbBlob, nil
}

func (crr *ResumeRequest) createOrUpdateModelBlob(ctx context.Context, mutateFn func(blob *model.ArtifactBlob) error) (*model.ArtifactBlob, error) {
	dbBlob, err := crr.artStore.Blobs().GetByRef(ctx, crr.Ref, crr.view.StorageID)
	if err == nil {
		return crr.artStore.Blobs().UpdateOptLock(ctx, dbBlob, mutateFn)
	}

	if errors.Is(err, pkgerrors.ErrResourceNotFound) {
		var creatorID = int64(0)
		session, ok := req.GetAuthSession(ctx)
		if ok {
			creatorID = session.User.ID
		}

		tm := time.UnixMilli(crr.State.StartedAt)
		_ = tm
		newObj := model.CreateArtifactBlob{
			StorageID: crr.view.StorageID,
			Ref:       crr.Ref,
			CreatedBy: fmt.Sprintf("%d", creatorID),
		}
		if err = mutateFn(&model.ArtifactBlob{StorageID: newObj.StorageID, Ref: newObj.Ref, CreatedBy: newObj.CreatedBy}); err != nil {
			return nil, err
		}
		// mutateFn sets Metadata and Size on a temp blob; we read from it.
		tmpBlob := &model.ArtifactBlob{StorageID: newObj.StorageID, Ref: newObj.Ref, CreatedBy: newObj.CreatedBy}
		if err = mutateFn(tmpBlob); err != nil {
			return nil, err
		}
		newObj.Metadata = tmpBlob.Metadata
		newObj.Size = tmpBlob.Size
		created, err := crr.artStore.Blobs().Create(ctx, &newObj)
		if err != nil {
			return nil, err
		}
		return created, err
	}
	return nil, err
}

func HashUUID(base uuid.UUID, salt string) string {
	return uuid.NewSHA1(base, []byte(salt)).String()
}

func parseContentRange(cr string) (start int64, end int64, err error) {
	rStart, rEnd, ok := strings.Cut(cr, "-")
	if !ok {
		return -1, -1, fmt.Errorf("invalid content range format, %s", cr)
	}
	start, err = strconv.ParseInt(rStart, 10, 64)
	if err != nil {
		return -1, -1, err
	}
	end, err = strconv.ParseInt(rEnd, 10, 64)
	if err != nil {
		return -1, -1, err
	}
	return start, end, nil
}

// The copy will be limited to `limit` bytes, if limit is greater than zero.
func copyFullPayload(ctx context.Context, r *http.Request, destWriter io.Writer, digester digest.Digester) error {
	// Get a channel that tells us if the client disconnects
	clientClosed := r.Context().Done()

	mw := io.MultiWriter(destWriter, digester.Hash())
	// Read in the data, if any.
	copied, err := io.Copy(mw, r.Body)
	if clientClosed != nil && (err != nil || (r.ContentLength > 0 && copied < r.ContentLength)) {
		// Didn't receive as much content as expected. Did the client
		// disconnect during the request? If so, avoid returning a 400
		// error to keep the logs cleaner.
		select {
		case <-clientClosed:
			return ErrClientClosed.WithDetail("client disconnected")
		default:
		}
	}

	if err != nil {
		logger.GetLogger(ctx).Errorf("unknown error reading request payload: %v", err)
		return ErrUnknown.WithDetail(err.Error())
	}

	return nil
}
