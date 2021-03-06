# IRISLCD用户文档

## 基本功能介绍

1.提供restful API以及swagger-ui
2.验证查询结果的有效性

## 使用场景

假设有一个正在运行的IRISLCD节点，其swagger-ui页面url是`localhost:1317/swagger-ui/`。IRISLCD用户文档的默认home目录是`$HOME/.irislcd`。IRISLCD启动以后，首先它将在其主文件夹中创建存储密钥的文件。如果IRISLCD运行在非信任模式下，它将获取当前区块作为其信任基础，并且将其保存到其home目录下的`trust-base.db`。IRISLCD节点始终无条件信任这个区块。它将根据这个区块的验证人集合来验证此后所有查询结果，这意味着IRISLCD只能验证其信任区块之后的区块信息和交易。这也是IRISLCD的缺陷。当它尝试验证较低高度的某些交易和区块时，会有错误返回。因此，如果您想查询早期的区块和交易时，请以信任模式启动IRISLCD。有关详细的查询结果的验证算法，请参阅此[文档](https://github.com/tendermint/tendermint/blob/master/docs/tendermint-core/light-client-protocol.md)。

IRISLCD节点成功启动后，在浏览器中打开`localhost1317/swagger-ui/`，您可以看到所有REST APIs。

1. Tendermint相关APIs, 例如查询区块，交易和验证人集
    1. `GET /node_info`: 查询所连接全节点的信息
    2. `GET /syncing`: 查询所连接全节点是否处于追赶区块的状态
    3. `GET /blocks/latest`: 获取最新区块
    4. `GET /blocks/{height}`: 获取某一高度的区块
    5. `GET /validatorsets/latest`: 获取最新的验证人集合
    6. `GET /validatorsets/{height}`: 获取某一高度的验证人集合
    7. `GET /txs/{hash}`: 通过交易hash查询交易
    8. `GET /txs`: 搜索交易
    9. `POST /txs`: 广播交易
 
2. Key management APIs

    1. `GET /keys`: 列出所有本地的秘钥
    2. `POST /keys`: 创建新的秘钥
    3. `GET /keys/seed`: 创建新的助记词
    4. `GET /keys/{name}`: 根据秘钥名称查询秘钥
    5. `PUT /keys/{name}`: 更新秘钥的密码
    6. `DELETE /keys/{name}`: 删除秘钥
    7. `GET /auth/accounts/{address}`: 查询秘钥对象账户的信息

3. Create, sign and broadcast transactions

    1. `POST /tx/sign`: 签名交易
    2. `POST /tx/broadcast`: 广播一个amino编码的交易
    3. `POST /txs/send`: 广播一个非amino编码的交易
    4. `GET /bank/coin/{coin-type}`: 查询coin的类型信息
    5. `GET /bank/balances/{address}`: 查询账户的token数量
    6. `POST /bank/accounts/{address}/transfers`: 发起转账交易

4. Stake module APIs

    1. `POST /stake/delegators/{delegatorAddr}/delegate`: 发起委托交易
    2. `POST /stake/delegators/{delegatorAddr}/redelegate`: 发起转委托交易
    3. `POST /stake/delegators/{delegatorAddr}/unbond`: 发起解委托交易
    4. `GET /stake/delegators/{delegatorAddr}/delegations`: 查询委托人的所有委托记录
    5. `GET /stake/delegators/{delegatorAddr}/unbonding_delegations`: 查询委托人的所有解委托记录
    6. `GET /stake/delegators/{delegatorAddr}/redelegations`: 查询委托人的所有转委托记录
    7. `GET /stake/delegators/{delegatorAddr}/validators`: 查询委托人的所委托的所有验证人
    8. `GET /stake/delegators/{delegatorAddr}/validators/{validatorAddr}`: 查询某个被委托的验证人上信息
    9. `GET /stake/delegators/{delegatorAddr}/txs`: 查询所有委托人相关的委托交易
    10. `GET /stake/delegators/{delegatorAddr}/delegations/{validatorAddr}`: 查询委托人在某个验证人上的委托记录
    11. `GET /stake/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr}`: 查询委托人在某个验证人上所有的解委托记录
    12. `GET /stake/validators`: 获取所有委托人信息
    13. `GET /stake/validators/{validatorAddr}`: 获取某个委托人信息
    14. `GET /stake/validators/{validatorAddr}/unbonding_delegations`: 获取某个验证人上的所有解委托记录
    15. `GET /stake/validators/{validatorAddr}/redelegations`: 获取某个验证人上的所有转委托记录
    16. `GET /stake/pool`: 获取权益池信息
    17. `GET /stake/parameters`: 获取权益证明的参数

5. Governance module APIs

    1. `POST /gov/proposal`: 发起提交提议交易
    2. `GET /gov/proposals`: 查询提议
    3. `POST /gov/proposals/{proposalId}/deposits`: 发起缴纳押金的交易
    4. `GET /gov/proposals/{proposalId}/deposits`: 查询缴纳的押金
    5. `POST /gov/proposals/{proposalId}/votes`: 发起投票交易
    6. `GET /gov/proposals/{proposalId}/votes`: 查询投票
    7. `GET /gov/proposals/{proposalId}`: 查询某个提议
    8. `GET /gov/proposals/{proposalId}/deposits/{depositor}`:查询押金
    9. `GET /gov/proposals/{proposalId}/votes/{voter}`: 查询投票
    10. `GET/gov/params`: 查询可供治理的参数

6. Slashing module APIs

    1. `GET /slashing/validators/{validatorPubKey}/signing_info`: 获取验证人的签名记录
    2. `POST /slashing/validators/{validatorAddr}/unjail`: 赦免某个作恶的验证人节点

7. Distribution module APIs

    1. `POST /distribution/{delegatorAddr}/withdrawAddress`: 设置收益取回地址
    2. `GET /distribution/{delegatorAddr}/withdrawAddress`: 查询收益取回地址
    3. `POST /distribution/{delegatorAddr}/withdrawReward`: 取回收益
    4. `GET /distribution/{delegatorAddr}/distrInfo/{validatorAddr}`: 查询某个委托的收益分配信息
    5. `GET /distribution/{delegatorAddr}/distrInfos`: 查询委托人所有委托的收益分配信息
    6. `GET /distribution/{validatorAddr}/valDistrInfo`: 查询验证人的收益分配信息

8. Query app version

    1. `GET /version`: 获取IRISHUB的版本
    2. `GET /node_version`: 查询全节点版本

## Options for post apis

1. `POST /bank/accounts/{address}/transfers`
2. `POST /stake/delegators/{delegatorAddr}/delegate`
3. `POST /stake/delegators/{delegatorAddr}/redelegate`
4. `POST /stake/delegators/{delegatorAddr}/unbond`
5. `POST /gov/proposal`
6. `POST /gov/proposals/{proposalId}/deposits`
7. `POST /gov/proposals/{proposalId}/votes`
8. `POST /slashing/validators/{validatorAddr}/unjail`

| 参数名字        | 类型 | 默认值 | 优先级 | 功能描述                 |
| --------------- | ---- | ------- |--------- |--------------------------- |
| generate-only   | bool | false | 0 | 构建一个未签名的交易并返回 |
| simulate        | bool | false | 1 | 用仿真的方式去执行交易 |
| async           | bool | false | 2 | 用异步地方式广播交易  |

以上八个APIs都有三个额外的查询参数，如上表所示。默认情况下，它们的值都是false。每个参数都有其唯一的优先级（这里`0`是最高优先级）。 如果指定了多个参数，则将忽略优先级较低的参数。 例如，如果`generate-only`为真，那么其他参数，例如`simulate`和`async`将被忽略。