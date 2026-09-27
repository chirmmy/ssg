---
title: CauseHSI项目说明
description: "Readme for CauseHSI: Cross-scene Hyperspectral image classifaction via Causal Disentanglement"
pubDate: 2025-07-21
Updated: 2025-07-21
tags: ["开源", "README"]
draft: false
---
# 1 简介
CauseHSI项目为论文CauseHSI: Cross-scene Hyperspectral image classifaction via Causal Disentanglement提供代码实现。CauseHSI用于高光谱图像跨场景分类任务，其架构图如下：

![image-20250724193644712](https://qiniu.anburger.site/post/image-20250724193644712.png)

CauseHSI主要分为两个部分：生成模块Generation Module(CGM)和因果解耦模块Causal Disentanglement Module(CDM)

# 2. 使用

## 2.1 准备

### 2.1.1 Requirements

```shell
pip intsall -r requirements.txt
```

### 2.1.2 数据集存储结构

```
#### 数据集目录结构
├── data
│   └── datasets
│       ├── Houston
│       ├── Pavia
│       └── HyRANK
└── ...
```

### 2.1.3 Pre-Augmented

在我们的方法中，需要对源域数据进行一次DCT变换和IDCT变换，我们的方法是基于Patch的高光谱图像分类任务，如果对每一个Patch进行变换降低了模型计算速度。因此，我们推荐您直接对原始数据（一张完整的高光谱图像）进行变换
```shell
python scg_aug.py
```

## 2.2 Train
- Houston:
```shell
python train-gan.py --data_path Houston/ --source_domain Houston13 --target_domain Houston18 \
  --training_sample_ratio 0.8 --flip_augmentation --radiation_augmentation
```
- Pavia:
```shell
python train-gan.py --data_path Pavia/ --source_domain PaviaU --target_domain PaviaC \
  --training_sample_ratio 0.5
```

- HyRANK:
```shell
python train-gan.py --data_path HyRANK/ --source_domain Dioni --target_domain Loukia \
  --training_sample_ratio 0.8
```

## 2.3 Visualization


## 2.4 Loss

### 2.4.1 Causal-inspired Disentanglement Loss
2.1. independence loss

2.2.1 reconstruction loss

2.2.2 causal loss for disentanglement


2.3. consistency loss


### 2.4.2 Domain Generation Loss
balance loss for Generator


### 2.4.3 Basic Loss

SD classification loss


ED classification loss




# 3. Acknowledgments
我们在此衷心感谢以下作者提供的开源代码：
- 我们的论文在数据处理、空间随机化和光谱随机化采用了xxx所提供的代码实现，并且该论文也是我们做该研究的启发
- 数据集来自于 https://github.com/YuxiangZhang-BIT/Data-CSHSI 整理的Houston、Pavia和HyRANK
- 特征提取模块我们采用了CVSS的部分结构
- 对于图像的采用DCT变换是常用手段，我们对于该部分的处理借鉴了自然图像领域的处理，主要是来自于论文xxx
- HSIC独立性评判准则是一个经典方法，我们采用了xxx所提供的代码实现，同时简单沿用了早期分支的思想。该论文也启发我们尝试使用解耦作为解决高光谱跨场景分类任务的解决方案，而不是继续延续现在最常用的对比学习和对比对抗学习
-
