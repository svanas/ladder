# Ladder

Ladder is crypto trading software that will incrementally buy or sell any crypto asset (rather than opting for a single price).

## reason to be

Trading crypto is a game of dealing with harsh volatility and the uncertainties of timing the market. Most traders get it wrong. Throw in human emotions and all you get is anxiety rather than profits. Ladder trading is a strategy to minimize the risks and reap greater rewards. This software automates the strategy while removing human emotions.

## what is latter trading?

Ladder trading is incremental buying or selling of any crypto asset rather than opting for a single price. These incremental buy or sell orders are called ladder steps. Ladder trading will limit your losses during the market fluctuations by spreading in and out of positions. 

Should you want to exit the market and sell your asset, your average sell price will be closer to the ATH. Should you want to enter the market and buy an asset, you'll be buying dips and your average buy price will be closer to the bottom. Either way, you'll maximize your profits by pushing the average price in the desired direction.

In essence, you are executing a better dollar-cost-averaging (DCA) trading strategy where you are selling (or buying) an asset regularly, doing away with the attempt at the market timing. Because this software will slowly but surely increase your order size, your average price will be better (compared to non-laddered DCA)

## benefits

* minimize the risks
* does away with the need for timing the market
* tackling high volatility
* less anxiety

## exchanges

At the time of this writing, the software supports the following exchanges:
* [Coinbase](https://www.coinbase.com)
* [Bitstamp](https://www.bitstamp.net)
* [Binance](https://www.binance.com)
* [Kraken](https://www.kraken.com)
* [ParaSwap](https://www.paraswap.io)
* [1inch v3](https://1inch.io) (we are working on v4)

## who is paying for gas?

Please note this software will NEVER broadcast an on-chain transaction to the network and spend your ETH on gas. This software only ever interacts with (off-chain) CEXes or send signatures (aka gasless transactions) to a DEX.

If on-chain transactions are needed (for example: to approve an asset or cancel your limit orders on a DEX), this software will prompt you with a nice message and be done with it. This is the designed behavior and not a bug.

When your limit orders are getting filled on a DEX, market makers are paying for the gas.

## usage

`./ladder [command] [flags]`

## commands

Use `./ladder [command] --help` for more information about a command.

Please note none of the below commands will actually place any orders unless you include `--dry-run=false` with your command line.

## sell

Usage: `./ladder sell [flags]`

| flag              | description                                                                | default |
|-------------------|----------------------------------------------------------------------------|---------|
| --exchange        | name or code of the exchange                                               |         |
| --asset           | name of the asset you will want to sell                                    | BTC     |
| --quote           | name of the asset you will want to receive                                 | USDT    |
| --start-at-price  | price where you will want to start selling at                              |         |
| --stop-at-price   | price where you will want to stop selling                                  |         |
| --start-with-size | size of your first sell order (in base asset)                              |         |
| --mult            | multiplier that defines the number of orders and the distance between them | 1.05    |
| --size            | the quantity you will want to sell (in base asset)                         |         |

## buy

Usage: `./ladder buy [flags]`

| flag              | description                                                                | default |
|-------------------|----------------------------------------------------------------------------|---------|
| --exchange        | name or code of the exchange                                               |         |
| --asset           | name of the asset you will want to buy                                     | BTC     |
| --quote           | name of the asset you will want to spend                                   | USDT    |
| --start-at-price  | price where you will want to start buying at                               |         |
| --stop-at-price   | price where you will want to stop buying                                   |         |
| --start-with-size | size of your first buy order (in quote asset)                              |         |
| --mult            | multiplier that defines the number of orders and the distance between them | 1.05    |
| --size            | the quantity you will want to buy (in quote asset)                         |         |

## cancel

Usage: `./ladder cancel [flags]`

| flag              | description                  |
|-------------------|------------------------------|
| --exchange        | name or code of the exchange |
| --asset           | base asset                   |
| --quote           | quote asset                  |
| --side            | `buy` or `sell`              |

## compiling

1. Download and install [Go version 1.21](https://go.dev) (or later)
2. Navigate to [Infura](https://www.infura.io) and generate yourself an API key
3. Open [infura.api.key](https://github.com/svanas/ladder/blob/main/api/web3/infura.api.key) and paste your Infura API key
4. Navigate to [CoinGecko](https://www.coingecko.com/en/developers/dashboard) and generate yourself an API key
5. Open [coingecko.api.key](https://github.com/svanas/ladder/blob/main/api/coingecko/coingecko.api.key) and paste your CoinGecko API key
6. Navigate to the [1inch dev portal](https://portal.1inch.dev) and generate yourself an API key
7. Open [1inch.api.key](https://github.com/svanas/ladder/blob/main/api/oneinch/1inch.api.key) and paste your [1inch](https://1inch.io) API key
8. Open a terminal in the directory you cloned this repo, and execute `go build` on the command-line
9. Should you decide to fork this repo, then do not commit your API keys. Your API keys are not to be shared.

## support

Ladder is free, unsupported software. But if you got questions, you are welcome to join [this Telegram group](https://t.me/laddercryptobot).

## disclaimer

Ladder is provided free of charge. There is no warranty. The authors do not assume any responsibility for bugs, vulnerabilities, or any other technical defects. Use at your own risk.
