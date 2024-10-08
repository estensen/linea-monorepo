package net.consensys.zkevm.coordinator.clients.prover

import net.consensys.linea.contract.parseBridgeLogsData
import org.web3j.protocol.core.methods.response.EthLog

object CommonTestData {
  val validTransactionRlp = this::class.java.getResource("/valid-transaction.rlp")!!.readText().trim()
  val ethLogs: List<EthLog.LogObject> =
    listOf(
      EthLog.LogObject(
        false,
        "0x64",
        "0x3c",
        "0xc2375a8e64bf9c3f4c2b1700735f6042a37e44d0c3df8c1735daeaa684197874",
        "0xb042425dedd752d905e97706ff4778d92fa6e3971017392380d8bef0153d0603",
        "0xb663a",
        "0x91ba8a14d2cc851abb69212c09f59e06e1e7f0a5",
        "0x0000000000000000000000000000000000000000000000008ac7230489" +
          "e800000000000000000000000000007ef911f8ef130f73d166468c0068753932357b17",
        null,
        listOf(
          "0x1cedfb451d4da3a63f72a945ba92f51b8fd558a5be4652404d037bb1578ff582",
          "0x0000000000000000000000000000000000000000000000000000000000040648"
        )
      ),
      EthLog.LogObject(
        false,
        "0x177",
        "0x8d",
        "0x4796d8934c88a1faa97aa6cf45a22aff130a51d0b8aeb0fbf5d7adb9985ea298",
        "0xb042425dedd752d905e97706ff4778d92fa6e3971017392380d8bef0153d0603",
        "0xb663a",
        "0x91ba8a14d2cc851abb69212c09f59e06e1e7f0a5",
        "0x",
        null,
        listOf(
          "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
          "0x0000000000000000000000000000000000000000000000000000000000000000",
          "0x000000000000000000000000f0bed1cff809d9314809a6905ad3d6efd7a062e7",
          "0x0000000000000000000000000000000000000000000000000000000000070341"
        )
      )
    )
  val bridgeLogs = ethLogs.map { ethLog -> parseBridgeLogsData(ethLog) }
}
