#include <string>
#include <iostream>
#include <fstream>
#include <sstream>
#include <memory>

using std::string;
using std::unique_ptr;
using std::ofstream;
using std::ostringstream;
using std::cin;

#include <glog/logging.h>
#include "tao/fd_message_channel.h"
#include "tao/tao_rpc.h"

using tao::FDMessageChannel;
using tao::InitializeApp;
using tao::Tao;
using tao::TaoRPC;

int main(int argc, char **argv) {
  InitializeApp(&argc, &argv, true);

  if (argc!=2) {
    LOG(FATAL) << "Invalid command line";
  }

  ostringstream buf_in;
  buf_in << cin.rdbuf();

  // This code expects fd 3 and 4 to be the pipes from and to the Tao, so it
  // doesn't need to take any parameters. It will establish a Tao Child Channel
  // directly with these fds.
  unique_ptr<FDMessageChannel> msg(new FDMessageChannel(3, 4));
  unique_ptr<Tao> tao(new TaoRPC(msg.release()));

  string buf_out;
  if (!tao->Seal(buf_in.str(), Tao::SealPolicyDefault, &buf_out)) {
    LOG(FATAL) << "Couldn't seal bytes across the channel";
  }

  ofstream fout(string("/storage/")+argv[1]);
  if (!fout.is_open()) {
    LOG(FATAL) << "Couldn't open file";
  }
  fout << buf_out;
  fout.close();

  return 0;
}
