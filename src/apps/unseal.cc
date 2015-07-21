#include <string>
#include <iostream>
#include <fstream>
#include <sstream>
#include <memory>

using std::string;
using std::unique_ptr;
using std::ifstream;
using std::ostringstream;
using std::cout;

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

  ifstream fin(string("/storage/")+argv[1]);
  if (!fin.is_open()) {
    LOG(FATAL) << "Couldn't open file";
  }
  ostringstream buf_in;
  buf_in << fin.rdbuf();
  fin.close();

  // This code expects fd 3 and 4 to be the pipes from and to the Tao, so it
  // doesn't need to take any parameters. It will establish a Tao Child Channel
  // directly with these fds.
  unique_ptr<FDMessageChannel> msg(new FDMessageChannel(3, 4));
  unique_ptr<Tao> tao(new TaoRPC(msg.release()));

  string buf_out;
  string policy;
  if (!tao->Unseal(buf_in.str(), &buf_out, &policy)) {
    LOG(FATAL) << "Couldn't seal bytes across the channel";
  }
  if (policy.compare(Tao::SealPolicyDefault) != 0) {
    LOG(FATAL) << "The policy returned by Unseal didn't match the Seal policy";
  }

  cout << buf_out;

  return 0;
}
