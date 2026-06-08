package usecase

import (
	"context"
	"os/exec"
)

const (
	wcaFFmpegBin = "ffmpeg"
)

func wcaFFmpegCmd(args ...string) *exec.Cmd {
	return exec.Command(wcaFFmpegBin, args...)
}

func wcaFFmpegCmdCtx(ctx context.Context, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, wcaFFmpegBin, args...)
}
