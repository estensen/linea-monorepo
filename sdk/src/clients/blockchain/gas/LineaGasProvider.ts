import { Block, JsonRpcProvider, TransactionReceipt, TransactionRequest, TransactionResponse } from "ethers";
import {
  LineaEstimateGasResponse,
  LineaGasFees,
  ILineaGasProvider,
  LineaGasProviderConfig,
} from "../../../core/clients/blockchain/IGasProvider";
import { IChainQuerier } from "../../../core/clients/blockchain/IChainQuerier";

const BASE_FEE_MULTIPLIER = 1.35;
const PRIORITY_FEE_MULTIPLIER = 1.05;

export class LineaGasProvider implements ILineaGasProvider<TransactionRequest> {
  /**
   * Creates an instance of `LineaGasProvider`.
   *
   * @param {IChainQuerier} chainQuerier - The chain querier for interacting with the blockchain.
   * @param {LineaGasProviderConfig} config - The configuration for the Linea gas provider.
   */
  constructor(
    protected readonly chainQuerier: IChainQuerier<
      TransactionReceipt,
      Block,
      TransactionRequest,
      TransactionResponse,
      JsonRpcProvider
    >,
    private readonly config: LineaGasProviderConfig,
  ) {}

  /**
   * Fetches gas fee estimates for a given transaction request.
   *
   * @param {TransactionRequest} transactionRequest - The transaction request to determine specific gas fees.
   * @returns {Promise<LineaGasFees>} A promise that resolves to an object containing Linea gas fee estimates.
   */
  public async getGasFees(transactionRequest: TransactionRequest): Promise<LineaGasFees> {
    const gasFees = await this.getLineaGasFees(transactionRequest);

    if (this.config.enforceMaxGasFee) {
      return {
        ...gasFees,
        maxPriorityFeePerGas: this.config.maxFeePerGas,
        maxFeePerGas: this.config.maxFeePerGas,
      };
    }
    return gasFees;
  }

  /**
   * Fetches Linea gas fee estimates for a given transaction request.
   *
   * @private
   * @param {TransactionRequest} transactionRequest - The transaction request to determine specific gas fees.
   * @returns {Promise<LineaGasFees>} A promise that resolves to an object containing Linea gas fee estimates.
   */
  private async getLineaGasFees(transactionRequest: TransactionRequest): Promise<LineaGasFees> {
    const lineaGasFees = await this.fetchLineaResponse(transactionRequest);

    const baseFee = this.getValueFromMultiplier(BigInt(lineaGasFees.baseFeePerGas), BASE_FEE_MULTIPLIER);
    const maxPriorityFeePerGas = this.getValueFromMultiplier(
      BigInt(lineaGasFees.priorityFeePerGas),
      PRIORITY_FEE_MULTIPLIER,
    );
    const maxFeePerGas = this.computeMaxFeePerGas(baseFee, maxPriorityFeePerGas);
    const gasLimit = BigInt(lineaGasFees.gasLimit);

    return {
      maxPriorityFeePerGas,
      maxFeePerGas,
      gasLimit,
    };
  }

  /**
   * Fetches the Linea gas fee response from the blockchain.
   *
   * @private
   * @param {TransactionRequest} transactionRequest - The transaction request to determine specific gas fees.
   * @returns {Promise<LineaEstimateGasResponse>} A promise that resolves to the Linea gas fee response.
   */
  private async fetchLineaResponse(transactionRequest: TransactionRequest): Promise<LineaEstimateGasResponse> {
    const params = {
      from: transactionRequest.from,
      to: transactionRequest.to,
      value: transactionRequest.value?.toString(),
      data: transactionRequest.data,
    };

    return this.chainQuerier.sendRequest("linea_estimateGas", [params]);
  }

  /**
   * Calculates a value based on a multiplier.
   *
   * @private
   * @param {bigint} value - The original value.
   * @param {number} multiplier - The multiplier to apply.
   * @returns {bigint} The calculated value.
   */
  private getValueFromMultiplier(value: bigint, multiplier: number): bigint {
    return (value * BigInt(multiplier * 100)) / 100n;
  }

  /**
   * Computes the maximum fee per gas.
   *
   * @private
   * @param {bigint} baseFee - The base fee per gas.
   * @param {bigint} priorityFee - The priority fee per gas.
   * @returns {bigint} The computed maximum fee per gas.
   */
  private computeMaxFeePerGas(baseFee: bigint, priorityFee: bigint): bigint {
    return baseFee + priorityFee;
  }

  /**
   * Gets the maximum fee per gas as configured.
   *
   * @returns {bigint} The maximum fee per gas.
   */
  public getMaxFeePerGas(): bigint {
    return this.config.maxFeePerGas;
  }
}
