import { Result } from '@chainlink/gauntlet-core'
import { logger, prompt } from '@chainlink/gauntlet-core/dist/utils'
import { SolanaCommand, TransactionResponse } from '@chainlink/gauntlet-solana'
import { PublicKey } from '@solana/web3.js'
import { MAX_TRANSACTION_BYTES, ORACLES_MAX_LENGTH } from '../../../../lib/constants'
import { CONTRACT_LIST, getContract } from '../../../../lib/contracts'
import { Protobuf } from '../../../../core/proto'
import { descriptor as OCR2Descriptor } from '../../../../core/proto/ocr2Proto'
import { getRDD } from '../../../../lib/rdd'
import { makeSharedSecretEncryptions, SharedSecretEncryptions } from '../../../../core/sharedSecretEncryptions'
import { durationToNanoseconds, Millisecond } from '../../../../core/time'
import { divideIntoChunks } from '../../../../core/utils'
import BN from 'bn.js'

export type Input = {
  deltaProgressNanoseconds: number
  deltaResendNanoseconds: number
  deltaRoundNanoseconds: number
  deltaGraceNanoseconds: number
  deltaStageNanoseconds: number
  rMax: number
  s: number[]
  offchainPublicKeys: string[]
  peerIds: string[]
  reportingPluginConfig: {
    alphaReportInfinite: boolean
    alphaReportPpb: number
    alphaAcceptInfinite: boolean
    alphaAcceptPpb: number
    deltaCNanoseconds: number
  }
  maxDurationQueryNanoseconds: number
  maxDurationObservationNanoseconds: number
  maxDurationReportNanoseconds: number
  maxDurationShouldAcceptFinalizedReportNanoseconds: number
  maxDurationShouldTransmitAcceptedReportNanoseconds: number
  configPublicKeys: string[]
}

export default class WriteOffchainConfig extends SolanaCommand {
  static id = 'ocr2:write_offchain_config'
  static category = CONTRACT_LIST.OCR_2

  static examples = [
    'yarn gauntlet ocr2:write_offchain_config --network=devnet --state=EPRYwrb1Dwi8VT5SutS4vYNdF8HqvE7QwvqeCCwHdVLC',
  ]

  constructor(flags, args) {
    super(flags, args)

    this.require(!!this.flags.state, 'Please provide flags with "state"')
    this.require(
      !!process.env.SECRET,
      'Please specify the Gauntlet secret words e.g. SECRET="awe fluke polygon tonic lilly acuity onyx debra bound gilbert wane"',
    )
  }

  static makeInputFromRDD = (rdd: any, stateAddress: string): Input => {
    const aggregator = rdd.contracts[stateAddress]
    const config = aggregator.config
    const aggregatorOperators: string[] = aggregator.oracles.map((o) => o.operator)
    const operatorsPublicKeys = aggregatorOperators.map((o) => rdd.operators[o].ocrOffchainPublicKey[0])
    const operatorsPeerIds = aggregatorOperators.map((o) => rdd.operators[o].peerId[0])
    const operatorConfigPublicKeys = aggregatorOperators.map((o) => rdd.operators[o].ocrConfigPublicKey[0])

    const input: Input = {
      deltaProgressNanoseconds: durationToNanoseconds(config.deltaProgress).toNumber(),
      deltaResendNanoseconds: durationToNanoseconds(config.deltaResend).toNumber(),
      deltaRoundNanoseconds: durationToNanoseconds(config.deltaRound).toNumber(),
      deltaGraceNanoseconds: durationToNanoseconds(config.deltaGrace).toNumber(),
      deltaStageNanoseconds: durationToNanoseconds(config.deltaStage).toNumber(),
      rMax: config.rMax,
      s: config.s,
      offchainPublicKeys: operatorsPublicKeys,
      peerIds: operatorsPeerIds,
      reportingPluginConfig: {
        alphaReportInfinite: config.reportingPluginConfig.alphaReportInfinite,
        alphaReportPpb: Number(config.reportingPluginConfig.alphaReportPpb),
        alphaAcceptInfinite: config.reportingPluginConfig.alphaAcceptInfinite,
        alphaAcceptPpb: Number(config.reportingPluginConfig.alphaAcceptPpb),
        deltaCNanoseconds: durationToNanoseconds(config.reportingPluginConfig.deltaC).toNumber(),
      },
      maxDurationQueryNanoseconds: durationToNanoseconds(config.maxDurationQuery).toNumber(),
      maxDurationObservationNanoseconds: durationToNanoseconds(config.maxDurationObservation).toNumber(),
      maxDurationReportNanoseconds: durationToNanoseconds(config.maxDurationReport).toNumber(),
      maxDurationShouldAcceptFinalizedReportNanoseconds: durationToNanoseconds(
        config.maxDurationShouldAcceptFinalizedReport,
      ).toNumber(),
      maxDurationShouldTransmitAcceptedReportNanoseconds: durationToNanoseconds(
        config.maxDurationShouldTransmitAcceptedReport,
      ).toNumber(),
      configPublicKeys: operatorConfigPublicKeys,
    }
    return input
  }

  makeInput = (userInput: any): Input => {
    // TODO: Some format validation for user input
    if (userInput) return userInput as Input
    const rdd = getRDD(this.flags.rdd)
    return WriteOffchainConfig.makeInputFromRDD(rdd, this.flags.state)
  }

  serializeOffchainConfig = async (input: Input): Promise<Buffer> => {
    const { configPublicKeys, ...validInput } = input
    const proto = new Protobuf({ descriptor: OCR2Descriptor })
    const reportingPluginConfigProto = proto.encode(
      'offchainreporting2_config.ReportingPluginConfig',
      validInput.reportingPluginConfig,
    )
    const sharedSecretEncryptions = await this.generateSecretEncryptions(configPublicKeys)
    const offchainConfig = {
      ...validInput,
      offchainPublicKeys: validInput.offchainPublicKeys.map((key) => Buffer.from(key, 'hex')),
      reportingPluginConfig: reportingPluginConfigProto,
      sharedSecretEncryptions,
    }
    return Buffer.from(proto.encode('offchainreporting2_config.OffchainConfigProto', offchainConfig))
  }

  // constructs a SharedSecretEncryptions from
  // a set of SharedSecretEncryptionPublicKeys, the sharedSecret, and a cryptographic randomness source
  generateSecretEncryptions = async (operatorsPublicKeys: string[]): Promise<SharedSecretEncryptions> => {
    const gauntletSecret = process.env.SECRET
    return makeSharedSecretEncryptions(gauntletSecret!, operatorsPublicKeys)
  }

  validateInput = (input: Input): boolean => {
    const _isNegative = (v: number): boolean => new BN(v).lt(new BN(0))
    const nonNegativeValues = [
      'deltaProgressNanoseconds',
      'deltaResendNanoseconds',
      'deltaRoundNanoseconds',
      'deltaGraceNanoseconds',
      'deltaStageNanoseconds',
      'maxDurationQueryNanoseconds',
      'maxDurationObservationNanoseconds',
      'maxDurationReportNanoseconds',
      'maxDurationShouldAcceptFinalizedReportNanoseconds',
      'maxDurationShouldTransmitAcceptedReportNanoseconds',
    ]
    for (let prop in nonNegativeValues) {
      if (_isNegative(input[prop])) throw new Error(`${prop} must be non-negative`)
    }
    const safeIntervalNanoseconds = new BN(200).mul(Millisecond).toNumber()
    if (input.deltaProgressNanoseconds < safeIntervalNanoseconds)
      throw new Error(
        `deltaProgressNanoseconds (${input.deltaProgressNanoseconds} ns)  is set below the resource exhaustion safe interval (${safeIntervalNanoseconds} ns)`,
      )
    if (input.deltaResendNanoseconds < safeIntervalNanoseconds)
      throw new Error(
        `deltaResendNanoseconds (${input.deltaResendNanoseconds} ns) is set below the resource exhaustion safe interval (${safeIntervalNanoseconds} ns)`,
      )

    if (input.deltaRoundNanoseconds < input.deltaProgressNanoseconds)
      throw new Error(
        `deltaRoundNanoseconds (${input.deltaRoundNanoseconds}) must be less than deltaProgressNanoseconds (${input.deltaProgressNanoseconds})`,
      )
    const sumMaxDurationsReportGeneration = new BN(input.maxDurationQueryNanoseconds)
      .add(new BN(input.maxDurationObservationNanoseconds))
      .add(new BN(input.maxDurationReportNanoseconds))

    if (sumMaxDurationsReportGeneration.gte(new BN(input.deltaProgressNanoseconds)))
      throw new Error(
        `sum of MaxDurationQuery/Observation/Report (${sumMaxDurationsReportGeneration}) must be less than deltaProgressNanoseconds (${input.deltaProgressNanoseconds})`,
      )

    if (input.rMax <= 0 || input.rMax >= 255)
      throw new Error(`rMax (${input.rMax}) must be greater than zero and less than 255`)

    if (input.s.length >= 1000) throw new Error(`Length of S (${input.s.length}) must be less than 1000`)
    for (let i = 0; i < input.s.length; i++) {
      const s = input.s[i]
      if (s < 0 || s > ORACLES_MAX_LENGTH)
        throw new Error(`S[${i}] (${s}) must be between 0 and Max Oracles (${ORACLES_MAX_LENGTH})`)
    }

    return true
  }

  execute = async () => {
    const ocr2 = getContract(CONTRACT_LIST.OCR_2, '')
    const address = ocr2.programId.toString()
    const program = this.loadProgram(ocr2.idl, address)

    const state = new PublicKey(this.flags.state)
    const owner = this.wallet.payer

    // Throws on invalid input
    const input = this.makeInput(this.flags.input)
    this.validateInput(input)

    // Check correct format OCR Keys
    const offchainConfig = await this.serializeOffchainConfig(input)

    logger.info(`Offchain config size: ${offchainConfig.byteLength}`)
    this.require(offchainConfig.byteLength < 4096, 'Offchain config must be lower than 4096 bytes')

    // There's a byte limit per transaction. Write the config in chunks
    const offchainConfigChunks = divideIntoChunks(offchainConfig, MAX_TRANSACTION_BYTES)
    if (offchainConfigChunks.length > 1) {
      logger.info(
        `Config size (${offchainConfig.byteLength} bytes) is bigger than transaction limit. It will be configured using ${offchainConfigChunks.length} transactions`,
      )
    }

    await prompt('Start writing offchain config?')

    const txs: string[] = []
    for (let i = 0; i < offchainConfigChunks.length; i++) {
      logger.loading(`Sending ${i + 1}/${offchainConfigChunks.length}...`)
      const tx = await program.rpc.writeOffchainConfig(offchainConfigChunks[i], {
        accounts: {
          state: state,
          authority: owner.publicKey,
        },
      })
      txs.push(tx)
    }
    logger.success(`Last tx Write offchain config set on tx ${txs[txs.length - 1]}`)

    return {
      responses: [
        {
          tx: this.wrapResponse(txs[txs.length - 1], state.toString(), { state: state.toString() }),
          contract: state.toString(),
        },
      ],
    } as Result<TransactionResponse>
  }
}
