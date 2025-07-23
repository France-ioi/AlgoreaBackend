package items

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/payloadstest"
	"github.com/France-ioi/AlgoreaBackend/v2/app/token"
	"github.com/France-ioi/AlgoreaBackend/v2/app/tokentest"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func Test_saveGradeRequestParsed_UnmarshalJSON(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("error")
	mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT public_key "+
		"FROM `platforms` JOIN items ON items.platform_id = platforms.id WHERE (items.id = ?) LIMIT 1") + "$").
		WithArgs(901756573345831409).WillReturnError(expectedError)

	r := saveGradeRequestParsed{
		store:     database.NewDataStore(db),
		publicKey: tokentest.AlgoreaPlatformPublicKeyParsed(),
	}
	assert.PanicsWithError(t, expectedError.Error(), func() {
		_ = r.UnmarshalJSON([]byte(fmt.Sprintf(`{"score_token": %q, "answer_token": %q}`,
			token.Generate(payloadstest.ScorePayloadFromGrader(), tokentest.AlgoreaPlatformPrivateKeyParsed()),
			token.Generate(payloadstest.AnswerPayloadFromAlgoreaPlatform(), tokentest.AlgoreaPlatformPrivateKeyParsed()),
		)))
	})
	assert.NoError(t, mock.ExpectationsWereMet())
}
