MjCheckTing

此算法是查询麻将听牌并算番数。

第一部分：生成胡牌表格 第二部分：查询听牌算番

参考思路： 1 https://github.com/esrrhs/majiang_algorithm/blob/master/hu.md 2 编程珠玑第一章文件排序

一般查表思路： 开始的时候我参考编程珠玑，设想穷举所有胡牌牌型，尽量压缩手牌表示方法。将胡的牌按照万筒条字牌的顺序采用位运算压缩，存放在一个很大的数字里。然后将所有的可能根据每种牌的数量分开存放，比如所有符合(万：2 筒：6 条：3 字：3)为一个集合，以(万：2 筒：6 条：3 字：3)做key存储，查找时只要将手牌转换成对应的Key查找到对应的集合(集合可以排过序然后二分查找对应胡牌)... 走到这里就没再深思了，因为这个穷举产生表格的过程耗时非常久，容错率非常低故而pass...

进阶思路： 请参考思路1的链接，其主要思路是将麻将分为两个花色既Normnal(万、筒、条)和Zi(字牌)，因为Normal牌可以有顺子(ABC)而Zi牌中只有刻子(AAA)和将(BB)。我通俗的理解为，你摸麻将其实在玩两个游戏，1 叫做普通牌"胡牌" 2 叫做字牌"胡牌" 1&&2 那么你胡了。再通俗一点，1 万牌胡 2 筒牌胡 3 条牌胡 4 字牌胡 1&&2&&3&&4 那么你胡了，