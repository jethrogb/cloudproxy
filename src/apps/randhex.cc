#include <string>
#include <iostream>
#include <sstream>
#include <memory>

using std::string;
using std::unique_ptr;
using std::istringstream;
using std::cout;

#include <glog/logging.h>
#include "tao/fd_message_channel.h"
#include "tao/tao_rpc.h"

using tao::FDMessageChannel;
using tao::InitializeApp;
using tao::Tao;
using tao::TaoRPC;

static char hchars[16]={'0','1','2','3','4','5','6','7','8','9','a','b','c','d','e','f'};

int main(int argc, char **argv) {
  InitializeApp(&argc, &argv, true);

  if (argc!=2) {
    LOG(FATAL) << "Invalid command line";
  }

  // This code expects fd 3 and 4 to be the pipes from and to the Tao, so it
  // doesn't need to take any parameters. It will establish a Tao Child Channel
  // directly with these fds.
  unique_ptr<FDMessageChannel> msg(new FDMessageChannel(3, 4));
  unique_ptr<Tao> tao(new TaoRPC(msg.release()));

  string buf;
  size_t len=0,i;
  istringstream(argv[1]) >> len;

  if (!tao->GetRandomBytes(len, &buf) || buf.size()!=len) {
    LOG(FATAL) << "Couldn't get random data from the Tao RPC channel";
  }

  buf.resize(len*2);
  for (i=len;i>0;i--) {
    buf[i*2-1]=hchars[((unsigned char)buf[i-1])&0xf];
    buf[i*2-2]=hchars[((unsigned char)buf[i-1]) >>4];
  }

  cout << buf;

  return 0;
}
