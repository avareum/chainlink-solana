import Initialize from './initialize'
import OCR2InitializeFlow from './initialize.flow'
import BeginOffchainConfig from './offchainConfig/begin'
import CommitOffchainConfig from './offchainConfig/commit'
import SetOffchainConfigFlow from './offchainConfig/setOffchainConfig.flow'
import WriteOffchainConfig from './offchainConfig/write'
import PayRemaining from './payRemaining'
import ReadState from './read'
import SetBilling from './setBilling'
import SetConfig from './setConfig'
import SetPayees from './setPayees'
import SetupFlow from './setup.dev.flow'
import SetupRDDFlow from './setup.dev.rdd.flow'
import SetValidatorConfig from './setValidatorConfig'
import Transmit from './transmit.dev'
import Inspection from './inspection'

export default [
  Initialize,
  OCR2InitializeFlow,
  SetBilling,
  PayRemaining,
  SetPayees,
  SetConfig,
  SetValidatorConfig,
  ReadState,
  SetOffchainConfigFlow,
  BeginOffchainConfig,
  WriteOffchainConfig,
  CommitOffchainConfig,
  // Inspection
  ...Inspection,
  // ONLY DEV
  Transmit,
  SetupFlow,
  SetupRDDFlow,
]
