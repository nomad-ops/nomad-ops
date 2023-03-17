package github

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	sshstd "golang.org/x/crypto/ssh"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"

	"github.com/nomad-ops/nomad-ops/backend/application"
	"github.com/nomad-ops/nomad-ops/backend/domain"
	"github.com/nomad-ops/nomad-ops/backend/utils/log"
)

type GitProvider struct {
	ctx      context.Context
	logger   log.Logger
	cfg      GitProviderConfig
	parser   application.JobParser
	repoLock sync.Mutex
	repos    map[string]*git.Repository
	keyRepo  application.KeyRepo
}

type GitProviderConfig struct {
	ReposDir string
}

func CreateGitProvider(ctx context.Context,
	logger log.Logger,
	cfg GitProviderConfig,
	parser application.JobParser,
	keyRepo application.KeyRepo) (*GitProvider, error) {

	t := &GitProvider{
		ctx:     ctx,
		logger:  logger,
		cfg:     cfg,
		parser:  parser,
		repos:   map[string]*git.Repository{},
		keyRepo: keyRepo,
	}

	return t, nil
}

func (g *GitProvider) FetchDesiredState(ctx context.Context, src *domain.Source) (*application.DesiredState, error) {
	g.repoLock.Lock()
	defer g.repoLock.Unlock()
	var auth transport.AuthMethod
	if src.DeployKeyName != "" {

		key, err := g.keyRepo.GetKey(ctx, src.DeployKeyName)
		if err != nil {
			g.logger.LogError(ctx, "Could not GetKey:%v", err)
			return nil, err
		}

		// TODO remove
		g.logger.LogInfo(ctx, "Key:%s", key.Value)

		publicKeys, err := ssh.NewPublicKeys("git", []byte(key.Value), "")
		if err != nil {
			g.logger.LogError(ctx, "Could not NewPublicKeys:%v", err)
			return nil, err
		}

		publicKeys.HostKeyCallback = sshstd.InsecureIgnoreHostKey()
		auth = publicKeys
	}

	repoDir := filepath.Join(g.cfg.ReposDir, fmt.Sprintf("%x", md5.Sum([]byte(src.URL))), path.Base(src.URL))
	g.logger.LogTrace(ctx, "RepoDir:%v", repoDir)
	var wt *git.Worktree
	gitInfo := application.GitInfo{}
	if repo, ok := g.repos[src.ID]; ok {
		var err error
		wt, err = repo.Worktree()
		if err != nil {
			return nil, err
		}
		g.logger.LogTrace(ctx, "Pulling...")
		err = wt.PullContext(ctx, &git.PullOptions{
			Auth:          auth,
			ReferenceName: plumbing.NewBranchReferenceName(src.Branch),
			SingleBranch:  true,
		})
		if err == git.NoErrAlreadyUpToDate {
			g.logger.LogTrace(ctx, "Already up to date")
		}
		if err != nil && err != git.NoErrAlreadyUpToDate {
			g.logger.LogError(ctx, "PullContext failed:%v", err)
			return nil, err
		}
		g.logger.LogTrace(ctx, "Getting last commit...")

		cIter, err := repo.Log(&git.LogOptions{})
		if err != nil {
			g.logger.LogError(ctx, "repo.Log failed:%v", err)
			return nil, err
		}
		c, err := cIter.Next()
		if err != nil {
			g.logger.LogError(ctx, "cIter.Next failed:%v", err)
			return nil, err
		}
		gitInfo.GitCommit = c.Hash.String()
		g.logger.LogTrace(ctx, "Getting last commit...%v", gitInfo.GitCommit)
	} else {

		repo, err := git.CloneContext(ctx, memory.NewStorage(), memfs.New(), &git.CloneOptions{
			URL: src.URL,
			// Depth:         1, https://github.com/go-git/go-git/issues/207
			NoCheckout:    false,
			Auth:          auth,
			Progress:      os.Stdout,
			SingleBranch:  true,
			ReferenceName: plumbing.NewBranchReferenceName(src.Branch),
		})
		if err != nil {
			g.logger.LogError(ctx, "Could not clone:%s - %v", src.URL, err)
			return nil, err
		}

		wt, err = repo.Worktree()
		if err != nil {
			g.logger.LogError(ctx, "repo.Worktree failed:%v", err)
			return nil, err
		}
		cIter, err := repo.Log(&git.LogOptions{})
		if err != nil {
			g.logger.LogError(ctx, "repo.Log failed:%v", err)
			return nil, err
		}
		c, err := cIter.Next()
		if err != nil {
			g.logger.LogError(ctx, "cIter.Next failed:%v", err)
			return nil, err
		}
		gitInfo.GitCommit = c.Hash.String()
		g.repos[src.ID] = repo
	}

	pathInfo, err := wt.Filesystem.Stat(src.Path)
	if err != nil {
		g.logger.LogError(ctx, "Could not stat Path in repo:%v - %v", src.Path, err)
		return nil, err
	}

	desiredState := &application.DesiredState{
		GitInfo: gitInfo,
		Jobs:    map[string]*application.JobInfo{},
	}

	if pathInfo.IsDir() {
		fileInfos, err := wt.Filesystem.ReadDir(src.Path)
		if err != nil {
			g.logger.LogError(ctx, "wt.Filesystem.ReadDir failed:%v", err)
			return nil, err
		}

		for _, file := range fileInfos {
			if !strings.HasSuffix(file.Name(), ".nomad") && !strings.HasSuffix(file.Name(), ".hcl") {
				g.logger.LogTrace(ctx, "ignoring file:%v", file.Name())
				continue
			}
			f, err := wt.Filesystem.Open(wt.Filesystem.Join(src.Path, file.Name()))
			if err != nil {
				return nil, err
			}

			jobData, err := io.ReadAll(f)
			if err != nil {
				return nil, err
			}

			j, err := g.parser.ParseJob(ctx, string(jobData))
			if err != nil {
				g.logger.LogError(ctx, "Could not parse JobFile:%v - %v", file.Name(), err)
				continue
			}
			j.GitInfo = gitInfo
			desiredState.Jobs[*j.Name] = j
		}
	} else {
		f, err := wt.Filesystem.Open(src.Path)
		if err != nil {
			g.logger.LogError(ctx, " wt.Filesystem.Open(*src.Path) failed:%v", err)
			return nil, err
		}

		jobData, err := io.ReadAll(f)
		if err != nil {
			g.logger.LogError(ctx, " wt.Filesystem.Open(*src.Path).ReadAll failed:%v", err)
			return nil, err
		}

		j, err := g.parser.ParseJob(ctx, string(jobData))
		if err != nil {
			g.logger.LogError(ctx, "Could not parse JobFile:%v - %v", src.Path, err)
			return nil, err
		}
		j.GitInfo = gitInfo
		desiredState.Jobs[*j.Name] = j
	}
	if g.logger.IsTraceEnabled(ctx) {
		g.logger.LogTrace(ctx, "desiredState...%v", log.ToJSONString(desiredState))
	}

	return desiredState, nil
}
