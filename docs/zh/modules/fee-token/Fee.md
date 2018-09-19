# 简介

手续费上限使用 --fee指定。gas是衡量交易需要消耗多少资源的单位。gas上限用--gas指定。gas上限太小时，不够交易需要的gas；手续费太低时，每一单位gas支付的手续费太低，验证人节点也不会执行这笔交易。手续费(最小单位)/gas应该大于等于2*10^10。推荐将gas上限设置为20000，手续费上限设置为400000000000000iris。需要多少gas就会花多少手续费，剩余的手续费会被退还。

## Fee

IRIS网络中的验证人为了维护自己节点的安全性和网络的正常运行，付出了不小的成本。因此，在IRIS网络中的各种交易都需要支付一定的手续费。各种交易中的--fee选项就是用来指定交易中支付的手续费的上限的参数。

## Gas

验证人节点处理不同的交易时需要的资源也不同，例如转账交易只需要进行的计算、查询和修改，而创建验证人需要进行的计算、查询和修改较多。Gas是用来衡量每个交易需要的资源多少的，下面给出部分操作需要的gas数量:

- 进行一次从数据库中读取数据的操作需要的gas数量为：10 + 读取数据长度（以bytes为单位）
- 进行一次往数据库中写入数据的操作需要的gas数量为：10 + 10*写入数据长度（以bytes为单位）
- 进行一次签名验证：100
- ....

一笔交易中各种操作需要的gas加起来就是这个交易需要的总共的gas。用户可以使用--gas参数给自己发起的交易设置这笔交易的gas上限，如果一笔交易使用的gas超过了用户设定的gas上限则交易就无法成功执行，用户的手续费不会被扣除。如果一笔交易使用的gas未超过用户设定的gas上限，则交易可能被成功的执行(还需要检查交易是否符合要求)。当交易被成功执行的时候，用户支付的手续费为: 手续费上限 * 实际消耗的gas/gas上限。

gas价格 = 手续费上限 / gas上限，代表用户为每个单位的资源消耗支付的手续费价格。

为了使用户支付的手续费维持在一个合理的水平，我们给gas价格设定了一个下限，2^(-8) iris/gas，gas价格低于此限制的交易不会被执行。

例子
```
    iriscli stake unbond complete --from=test --address-validator=faa1mahw6ymzvt2q3lu4pjj5pau2e8krntklgarrxy --address-delegator=faa1mahw6ymzvt2q3lu4pjj5pau2e8krntklgarrxy  --fee=2000000000000000iris --gas=20000 --chain-id=test
```

在这个例子中执行的是完成解绑操作，这里设定的手续费上限(--fee)为2000000000000000iris(2*10^15)，gas上限(--gas)为20000，gas价格就是10^11iris/gas。假设执行交易总共需要1500个gas，那么会有1500000000000000iris的手续费被支付给验证人节点；剩余的500000000000000iris会被退还给用户。