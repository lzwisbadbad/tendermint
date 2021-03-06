package proxy

import (
	"github.com/bcbchain/tendermint/lite"
	certclient "github.com/bcbchain/tendermint/lite/client"
	"github.com/bcbchain/tendermint/lite/files"
)

func GetCertifier(chainID, rootDir, nodeAddr string) (*lite.InquiringCertifier, error) {
	trust := lite.NewCacheProvider(
		lite.NewMemStoreProvider(),
		files.NewProvider(rootDir),
	)

	source := certclient.NewHTTPProvider(nodeAddr)

	// XXX: total insecure hack to avoid `init`
	fc, err := source.LatestCommit()
	/* XXX
	// this gets the most recent verified commit
	fc, err := trust.LatestCommit()
	if certerr.IsCommitNotFoundErr(err) {
		return nil, errors.New("Please run init first to establish a root of trust")
	}*/
	if err != nil {
		return nil, err
	}

	cert, err := lite.NewInquiringCertifier(chainID, fc, trust, source)
	if err != nil {
		return nil, err
	}

	return cert, nil
}
