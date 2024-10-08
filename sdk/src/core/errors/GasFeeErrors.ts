import { MessageProps } from "../../core/entities/Message";
import { BaseError } from "./Base";

export class FeeEstimationError extends BaseError {
  override name = FeeEstimationError.name;
}

export class GasEstimationError<T extends MessageProps> extends BaseError {
  override name = GasEstimationError.name;
  public rejectedMessage?: T;

  constructor(message: string, rejectedMessage?: T) {
    super(message);
    this.rejectedMessage = rejectedMessage;
  }
}
