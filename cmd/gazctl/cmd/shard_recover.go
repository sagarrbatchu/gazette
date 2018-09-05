package cmd

import (
	"context"
	"encoding/json"
	"os"
	"path"

	"github.com/LiveRamp/gazette/pkg/consumer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/LiveRamp/gazette/pkg/gazette"
	"github.com/LiveRamp/gazette/pkg/recoverylog"

	rocks "github.com/tecbot/gorocksdb"
)

var shardCmd = &cobra.Command{
	Use:   "shard",
	Short: "Commands for working with a gazette consumer shard",
}

var shardRecoverCmd = &cobra.Command{
	Use:   "recover [hints-locator] [local-output-path]",
	Short: "Recover contents of the indicated log hints.",
	Long: `Recover replays the recoverylog indicated by the argument hints
into a local output directory, and additionally writes recovered JSON hints
upon completion.

hints-locator may be a path to a local file ("path/to/hints") or an Etcd
key specifier ("etcd:///path/to/hints").
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			cmd.Usage()
			log.Fatal("invalid arguments")
		}
		var hints, tgtPath = loadHints(args[0]), args[1]

		log.Info("running test version 5 (instrument normalization)")

		f, err := os.OpenFile("shard_recover_" + args[1] + ".log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		log.SetOutput(f)

		player, err := recoverylog.NewPlayer(hints, tgtPath)
		if err != nil {
			log.WithField("err", err).Fatal("preparing playback")
		}

		writeService := writeService()
		go func() {

			if err := player.PlayContext(
				context.Background(),
				struct {
					*gazette.Client
					*gazette.WriteService
				}{gazetteClient(), writeService},
			); err != nil {
				log.WithField("err", err).Fatal("shard playback failed")
			}
		}()

		var fsm = player.FinishAtWriteHead()
		if fsm == nil {
			return
		}

		// Write recovered hints under |tgtPath|.
		var recoveredPath = path.Join(tgtPath, "recoveredHints.json")

		fout, err := os.Create(recoveredPath)
		if err == nil {
			err = json.NewEncoder(fout).Encode(fsm.BuildHints())
		}
		if err != nil {
			log.WithFields(log.Fields{"err": err, "path": recoveredPath}).Fatal("failed to write recovered hints")
		} else {
			log.WithField("path", recoveredPath).Info("wrote recovered hints")
		}

		author, err := recoverylog.NewRandomAuthorID()
		log.WithField("author", author).Info("starting up recorder")
		var recorder = recoverylog.NewRecorder(fsm, author, len(tgtPath), writeService)
		recorder.BuildHints()
		log.WithField("recorder", recorder).Info("recorder initialized")

		var opts = rocks.NewDefaultOptions()
		if plugin := consumerPlugin(); plugin != nil {
			if initer, _ := plugin.(consumer.OptionsIniter); initer != nil {
				initer.InitOptions(opts)
			}
		}
		log.WithField("fsm", fsm).Info("opening consumer db")
		if newdb, err := consumer.NewDatabase(opts, fsm, author, tgtPath, writeService); err != nil {
			log.WithFields(log.Fields{"hints": args[0], "err": err}).Error("failed to open database")
		} else {
			println(newdb.GetLiveFilesMetaData())
			log.Info(newdb.GetLiveFilesMetaData())
			println("new db inited")
			log.Info("new db inited")
			newdb.Close()
			println("new db closed")
			log.Info("new db closed")
		}

		//// Let the consumer and runner perform any desired initialization or teardown.
		//if initer, ok := runner.Consumer.(ShardIniter); ok {
		//	if err = initer.InitShard(m); err != nil {
		//		log.WithFields(log.Fields{"shard": m.shard, "err": err}).Error("failed to InitShard")
		//		return
		//	}
		//}
		//
	},
}

func init() {
	rootCmd.AddCommand(shardCmd)

	shardCmd.AddCommand(shardRecoverCmd)
}
