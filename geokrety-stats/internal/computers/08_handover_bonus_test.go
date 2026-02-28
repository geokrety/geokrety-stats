package computers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/geokrety/geokrety-points-system/internal/computers"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
)

func TestHandoverBonus_AwardsOwnerOnUserToUserGrab(t *testing.T) {
	module := computers.NewHandoverBonus(newSpyStore(), testCfg())
	evt := testEvent(200, 4443, pipeline.LogTypeGrab)
	pipeCtx := testCtx(evt, 405, 1.0)
	pipeCtx.GKState.GKType = 0
	pipeCtx.GKState.CurrentHolder = 111
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)
	require.Len(t, acc.Awards(), 1)
	assert.Equal(t, 1.0, mustAwardByLabel(acc, "handover_owner").Points)
	assert.Equal(t, int64(405), mustAwardByLabel(acc, "handover_owner").RecipientUserID)
}

func TestHandoverBonus_SkipsWhenNotEligible(t *testing.T) {
	module := computers.NewHandoverBonus(newSpyStore(), testCfg())
	cases := []struct {
		name      string
		setupFunc func(*pipeline.Context)
	}{
		{
			name: "non-grab",
			setupFunc: func(p *pipeline.Context) { p.Event.LogType = pipeline.LogTypeDrop },
		},
		{
			name: "non-transferable",
			setupFunc: func(p *pipeline.Context) { p.GKState.GKType = 8 },
		},
		{
			name: "grabber is owner",
			setupFunc: func(p *pipeline.Context) { p.Event.UserID = p.GKState.OwnerID },
		},
		{
			name: "from cache",
			setupFunc: func(p *pipeline.Context) { p.GKState.CurrentHolder = 0 },
		},
		{
			name: "previous holder owner",
			setupFunc: func(p *pipeline.Context) { p.GKState.CurrentHolder = p.GKState.OwnerID },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			evt := testEvent(200, 4443, pipeline.LogTypeGrab)
			pipeCtx := testCtx(evt, 405, 1.0)
			pipeCtx.GKState.GKType = 0
			pipeCtx.GKState.CurrentHolder = 111
			tc.setupFunc(pipeCtx)
			acc := pipeline.NewAccumulator()

			require.NoError(t, module.Process(context.Background(), pipeCtx, acc))
			assert.Len(t, acc.Awards(), 0)
		})
	}
}
