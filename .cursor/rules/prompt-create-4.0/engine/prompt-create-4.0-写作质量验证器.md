---
alwaysApply: true
---

# 🔍 Prompt-Create-4.0-写作质量验证器

## 🎯 模块核心定位

### 设计理念：双平台写作质量智能验证与优化建议
> **作为4.0版本的质量保证核心，通过"内容质量验证 + 平台适配验证 + 用户体验验证 + 效果预测评估"四大机制，确保微信公众号与小红书双平台写作内容的高质量输出**

## 🧠 核心架构：双平台写作质量验证系统

```mermaid
graph TD
    A[写作创意内容] --> B[🔍 写作质量验证器 4.0]
    
    subgraph "四大验证机制"
        C1[📝 内容质量验证器<br/>语言质量+逻辑结构+价值密度]
        C2[📱 平台适配验证器<br/>平台特色+算法友好+用户匹配]
        C3[👥 用户体验验证器<br/>可读性+吸引力+互动性+实用性]
        C4[📊 效果预测评估器<br/>传播潜力+转化预期+商业价值]
    end
    
    subgraph "双平台验证标准"
        D1[📱 微信公众号标准<br/>专业性+深度性+权威性+价值性]
        D2[🌸 小红书标准<br/>真实性+生活化+情感化+种草力]
        D3[🔄 通用质量标准<br/>原创性+准确性+完整性+创新性]
    end
    
    subgraph "三级质量评定"
        E1[🏆 优秀级 (90-100分)<br/>平台爆款潜力]
        E2[✅ 良好级 (75-89分)<br/>平台标准达标]
        E3[⚠️ 待优化级 (<75分)<br/>需要改进优化]
    end
    
    B --> C1 --> C2 --> C3 --> C4
    
    C2 --> D1
    C2 --> D2
    C2 --> D3
    
    C4 --> E1
    C4 --> E2
    C4 --> E3
    
    E3 --> F[🎊 质量验证报告<br/>质量评分+问题诊断<br/>优化建议+改进方案]
```

## 💎 四大验证机制详解

### 📝 内容质量验证器
```yaml
核心功能:
  语言质量检测: ["语法准确性", "表达流畅性", "用词精准性", "风格统一性"]
  逻辑结构验证: ["逻辑清晰性", "结构完整性", "论证合理性", "层次分明性"]
  价值密度评估: ["信息价值", "实用价值", "思考价值", "情感价值"]
  原创性检测: ["内容原创度", "观点独特性", "表达创新性", "素材新颖性"]

内容质量算法:
  ```python
  def content_quality_validation(writing_content, platform_type):
      """内容质量验证算法"""
      # Step 1: 语言质量检测
      language_quality = analyze_language_quality(writing_content)
      
      # Step 2: 逻辑结构验证
      logical_structure = validate_logical_structure(writing_content)
      
      # Step 3: 价值密度评估
      value_density = assess_value_density(
          writing_content,
          platform_type
      )
      
      # Step 4: 原创性检测
      originality_check = check_content_originality(writing_content)
      
      # Step 5: 综合质量评分
      quality_score = calculate_content_quality_score(
          language_quality,
          logical_structure,
          value_density,
          originality_check
      )
      
      return {
          "language_quality": language_quality,
          "logical_structure": logical_structure,
          "value_density": value_density,
          "originality_check": originality_check,
          "quality_score": quality_score,
          "improvement_suggestions": generate_content_improvements(
              language_quality, logical_structure, value_density, originality_check
          )
      }
  ```

验证标准:
  语言质量标准:
    - 语法准确率: ≥ 98%
    - 表达流畅度: ≥ 90%
    - 用词精准度: ≥ 85%
    - 风格一致性: ≥ 88%
  
  逻辑结构标准:
    - 逻辑清晰度: ≥ 90%
    - 结构完整性: ≥ 95%
    - 论证合理性: ≥ 85%
    - 层次分明性: ≥ 88%
  
  价值密度标准:
    - 信息价值度: ≥ 80%
    - 实用价值度: ≥ 85%
    - 思考价值度: ≥ 75%
    - 情感价值度: ≥ 70%
```

### 📱 平台适配验证器
```yaml
核心功能:
  平台特色匹配: ["内容风格", "表达方式", "互动设计", "视觉呈现"]
  算法友好度检测: ["SEO优化", "关键词密度", "标题吸引力", "结构优化"]
  用户群体匹配: ["目标用户", "阅读习惯", "兴趣偏好", "消费能力"]
  传播机制适配: ["分享动机", "传播路径", "互动机制", "病毒传播"]

平台适配算法:
  ```python
  def platform_adaptation_validation(writing_content, platform_specifications):
      """平台适配验证算法"""
      # Step 1: 平台特色匹配检测
      platform_match = analyze_platform_characteristic_match(
          writing_content,
          platform_specifications
      )
      
      # Step 2: 算法友好度检测
      algorithm_friendly = assess_algorithm_friendliness(
          writing_content,
          platform_specifications
      )
      
      # Step 3: 用户群体匹配验证
      user_match = validate_target_user_match(
          writing_content,
          platform_specifications
      )
      
      # Step 4: 传播机制适配检测
      propagation_match = assess_propagation_mechanism_match(
          writing_content,
          platform_specifications
      )
      
      # Step 5: 综合适配评分
      adaptation_score = calculate_platform_adaptation_score(
          platform_match,
          algorithm_friendly,
          user_match,
          propagation_match
      )
      
      return {
          "platform_match": platform_match,
          "algorithm_friendly": algorithm_friendly,
          "user_match": user_match,
          "propagation_match": propagation_match,
          "adaptation_score": adaptation_score,
          "platform_optimization_suggestions": generate_platform_optimizations(
              platform_match, algorithm_friendly, user_match, propagation_match
          )
      }
  ```

平台适配标准:
  微信公众号适配标准:
    - 专业深度匹配: ≥ 85%
    - 知识价值密度: ≥ 80%
    - 商业价值潜力: ≥ 75%
    - 分享转发潜力: ≥ 70%
  
  小红书适配标准:
    - 生活化表达度: ≥ 85%
    - 真实体验感: ≥ 90%
    - 情感共鸣度: ≥ 80%
    - 种草转化力: ≥ 75%
  
  通用适配标准:
    - SEO友好度: ≥ 80%
    - 用户匹配度: ≥ 85%
    - 传播适配度: ≥ 75%
    - 算法推荐度: ≥ 70%
```

### 👥 用户体验验证器
```yaml
核心功能:
  可读性检测: ["阅读难度", "理解门槛", "阅读流畅性", "认知负荷"]
  吸引力评估: ["标题吸引力", "开头吸引力", "内容吸引力", "视觉吸引力"]
  互动性验证: ["互动设计", "参与门槛", "互动引导", "社交传播"]
  实用性评估: ["实用价值", "操作性", "可应用性", "问题解决"]

用户体验算法:
  ```python
  def user_experience_validation(writing_content, target_users):
      """用户体验验证算法"""
      # Step 1: 可读性检测
      readability_analysis = analyze_content_readability(
          writing_content,
          target_users
      )
      
      # Step 2: 吸引力评估
      attractiveness_assessment = assess_content_attractiveness(
          writing_content,
          target_users
      )
      
      # Step 3: 互动性验证
      interactivity_validation = validate_content_interactivity(
          writing_content,
          target_users
      )
      
      # Step 4: 实用性评估
      practicality_assessment = assess_content_practicality(
          writing_content,
          target_users
      )
      
      # Step 5: 综合体验评分
      experience_score = calculate_user_experience_score(
          readability_analysis,
          attractiveness_assessment,
          interactivity_validation,
          practicality_assessment
      )
      
      return {
          "readability_analysis": readability_analysis,
          "attractiveness_assessment": attractiveness_assessment,
          "interactivity_validation": interactivity_validation,
          "practicality_assessment": practicality_assessment,
          "experience_score": experience_score,
          "user_experience_optimizations": generate_ux_optimizations(
              readability_analysis, attractiveness_assessment, 
              interactivity_validation, practicality_assessment
          )
      }
  ```

用户体验标准:
  可读性标准:
    - 阅读难度适中: ≥ 85%
    - 理解门槛合理: ≥ 80%
    - 阅读流畅度: ≥ 90%
    - 认知负荷适当: ≥ 85%
  
  吸引力标准:
    - 标题吸引力: ≥ 80%
    - 开头吸引力: ≥ 85%
    - 内容吸引力: ≥ 80%
    - 视觉吸引力: ≥ 75%
  
  互动性标准:
    - 互动设计合理: ≥ 80%
    - 参与门槛适中: ≥ 85%
    - 互动引导清晰: ≥ 80%
    - 社交传播性: ≥ 75%
```

### 📊 效果预测评估器
```yaml
核心功能:
  传播潜力预测: ["阅读量预测", "分享率预测", "互动率预测", "传播路径分析"]
  转化预期评估: ["关注转化", "行为转化", "商业转化", "品牌转化"]
  商业价值评估: ["流量价值", "品牌价值", "销售价值", "长期价值"]
  风险评估: ["内容风险", "传播风险", "品牌风险", "合规风险"]

效果预测算法:
  ```python
  def effect_prediction_assessment(writing_content, platform_data, market_context):
      """效果预测评估算法"""
      # Step 1: 传播潜力预测
      propagation_potential = predict_propagation_potential(
          writing_content,
          platform_data,
          market_context
      )
      
      # Step 2: 转化预期评估
      conversion_expectation = assess_conversion_expectation(
          writing_content,
          platform_data,
          market_context
      )
      
      # Step 3: 商业价值评估
      business_value = assess_business_value_potential(
          writing_content,
          platform_data,
          market_context
      )
      
      # Step 4: 风险评估
      risk_assessment = assess_content_risks(
          writing_content,
          platform_data,
          market_context
      )
      
      # Step 5: 综合效果预测
      effect_prediction = generate_comprehensive_effect_prediction(
          propagation_potential,
          conversion_expectation,
          business_value,
          risk_assessment
      )
      
      return {
          "propagation_potential": propagation_potential,
          "conversion_expectation": conversion_expectation,
          "business_value": business_value,
          "risk_assessment": risk_assessment,
          "effect_prediction": effect_prediction,
          "optimization_recommendations": generate_effect_optimizations(
              propagation_potential, conversion_expectation, business_value, risk_assessment
          )
      }
  ```

效果预测标准:
  传播潜力标准:
    - 阅读量预测准确率: ≥ 80%
    - 分享率预测准确率: ≥ 75%
    - 互动率预测准确率: ≥ 70%
    - 传播路径预测准确率: ≥ 75%
  
  转化预期标准:
    - 关注转化预测: ≥ 70%
    - 行为转化预测: ≥ 65%
    - 商业转化预测: ≥ 60%
    - 品牌转化预测: ≥ 70%
  
  商业价值标准:
    - 流量价值评估: ≥ 75%
    - 品牌价值评估: ≥ 70%
    - 销售价值评估: ≥ 65%
    - 长期价值评估: ≥ 70%
```

## 🎯 双平台验证标准详解

### 📱 微信公众号验证标准
```yaml
专业性验证标准:
  内容专业度:
    - 行业知识准确性: ≥ 95%
    - 专业术语使用正确性: ≥ 90%
    - 数据引用准确性: ≥ 95%
    - 观点论证逻辑性: ≥ 85%
  
  表达专业度:
    - 语言表达严谨性: ≥ 90%
    - 逻辑结构清晰性: ≥ 88%
    - 论证过程完整性: ≥ 85%
    - 结论表达准确性: ≥ 90%

深度性验证标准:
  内容深度:
    - 思考深度: ≥ 80%
    - 分析深度: ≥ 85%
    - 见解独特性: ≥ 75%
    - 价值挖掘深度: ≥ 80%
  
  结构深度:
    - 层次结构深度: ≥ 85%
    - 论证层次丰富性: ≥ 80%
    - 案例分析深度: ≥ 85%
    - 总结提升深度: ≥ 80%

权威性验证标准:
  信息权威性:
    - 信息来源权威性: ≥ 90%
    - 引用资料可靠性: ≥ 95%
    - 数据统计权威性: ≥ 90%
    - 专家观点可信度: ≥ 85%
  
  表达权威性:
    - 观点表达自信度: ≥ 85%
    - 结论判断准确性: ≥ 90%
    - 建议实用性: ≥ 85%
    - 预测合理性: ≥ 80%

价值性验证标准:
  实用价值:
    - 实际应用性: ≥ 85%
    - 问题解决能力: ≥ 80%
    - 方法可操作性: ≥ 85%
    - 效果可预期性: ≥ 80%
  
  认知价值:
    - 认知提升度: ≥ 85%
    - 思维启发性: ≥ 80%
    - 知识增量: ≥ 85%
    - 视野拓展度: ≥ 80%
```

### 🌸 小红书验证标准
```yaml
真实性验证标准:
  体验真实性:
    - 个人体验真实度: ≥ 95%
    - 使用感受真实度: ≥ 90%
    - 效果描述真实度: ≥ 95%
    - 问题反馈真实度: ≥ 90%
  
  情感真实性:
    - 情感表达真实度: ≥ 90%
    - 情感变化真实度: ≥ 85%
    - 情感共鸣真实度: ≥ 85%
    - 情感传递真实度: ≥ 80%

生活化验证标准:
  表达生活化:
    - 语言生活化程度: ≥ 85%
    - 场景生活化程度: ≥ 90%
    - 案例生活化程度: ≥ 85%
    - 建议生活化程度: ≥ 80%
  
  内容生活化:
    - 话题生活化程度: ≥ 85%
    - 角度生活化程度: ≥ 80%
    - 价值生活化程度: ≥ 85%
    - 应用生活化程度: ≥ 80%

情感化验证标准:
  情感表达:
    - 情感丰富度: ≥ 80%
    - 情感层次度: ≥ 75%
    - 情感感染力: ≥ 85%
    - 情感持续性: ≥ 80%
  
  情感共鸣:
    - 用户情感共鸣度: ≥ 85%
    - 情感传播力: ≥ 80%
    - 情感记忆点: ≥ 75%
    - 情感行动力: ≥ 80%

种草力验证标准:
  种草内容:
    - 产品介绍吸引力: ≥ 85%
    - 使用效果说服力: ≥ 80%
    - 购买理由充分性: ≥ 85%
    - 价值感知度: ≥ 80%
  
  种草效果:
    - 购买欲望激发度: ≥ 80%
    - 行动转化可能性: ≥ 75%
    - 推荐传播意愿: ≥ 80%
    - 品牌认知提升: ≥ 75%
```

### 🔄 通用质量标准
```yaml
原创性标准:
  内容原创性:
    - 核心观点原创度: ≥ 80%
    - 表达方式原创度: ≥ 75%
    - 案例素材原创度: ≥ 85%
    - 结构设计原创度: ≥ 70%
  
  创新性:
    - 思维创新度: ≥ 75%
    - 角度创新度: ≥ 80%
    - 形式创新度: ≥ 70%
    - 价值创新度: ≥ 75%

准确性标准:
  事实准确性:
    - 数据准确率: ≥ 95%
    - 信息准确率: ≥ 95%
    - 引用准确率: ≥ 95%
    - 描述准确率: ≥ 90%
  
  逻辑准确性:
    - 推理逻辑准确性: ≥ 90%
    - 因果关系准确性: ≥ 85%
    - 结论逻辑准确性: ≥ 90%
    - 论证逻辑准确性: ≥ 85%

完整性标准:
  内容完整性:
    - 信息完整度: ≥ 85%
    - 论证完整度: ≥ 85%
    - 结构完整度: ≥ 90%
    - 价值完整度: ≥ 80%
  
  体验完整性:
    - 阅读体验完整性: ≥ 85%
    - 理解体验完整性: ≥ 85%
    - 价值体验完整性: ≥ 80%
    - 互动体验完整性: ≥ 75%

创新性标准:
  内容创新:
    - 主题创新度: ≥ 75%
    - 角度创新度: ≥ 80%
    - 方法创新度: ≥ 70%
    - 价值创新度: ≥ 75%
  
  表达创新:
    - 语言创新度: ≥ 70%
    - 结构创新度: ≥ 75%
    - 形式创新度: ≥ 70%
    - 互动创新度: ≥ 70%
```

## 🎨 三级质量评定详解

### 🏆 优秀级 (90-100分)
```yaml
评定标准:
  综合质量分: 90-100分
  各维度达标: 所有维度≥85分
  特色表现: 至少2个维度≥95分
  创新突破: 至少1个维度有明显创新突破

特征表现:
  - 平台爆款潜力: 具备成为平台爆款内容的潜力
  - 用户价值突出: 为用户提供超预期的价值
  - 传播价值显著: 具备强大的自然传播能力
  - 商业价值明确: 具备明确的商业变现潜力

获得建议:
  - 内容策略: 可作为内容标杆和模板
  - 传播策略: 可重点推广和传播
  - 商业策略: 可重点商业化运作
  - 复制策略: 可复制成功模式

优化建议:
  - 细节完善: 在已有优势基础上进一步完善细节
  - 效果放大: 通过推广策略放大内容效果
  - 模式总结: 总结成功经验形成可复制模式
  - 持续迭代: 在成功基础上持续创新迭代
```

### ✅ 良好级 (75-89分)
```yaml
评定标准:
  综合质量分: 75-89分
  各维度达标: 80%维度≥75分
  平台适配: 平台适配度≥80分
  用户体验: 用户体验≥75分

特征表现:
  - 平台标准达标: 符合平台发布标准和质量要求
  - 用户价值良好: 为用户提供有价值的内容
  - 传播潜力良好: 具备一定的传播能力
  - 商业价值适中: 具备基本的商业价值

获得建议:
  - 内容策略: 可正常发布和推广
  - 传播策略: 可进行常规传播和推广
  - 商业策略: 可进行适度商业化运作
  - 优化策略: 针对薄弱环节进行优化

优化建议:
  - 强化优势: 进一步强化已有优势维度
  - 补齐短板: 重点提升薄弱维度表现
  - 特色突出: 突出平台特色和个人特色
  - 效果提升: 通过优化提升整体效果
```

### ⚠️ 待优化级 (<75分)
```yaml
评定标准:
  综合质量分: <75分
  问题维度: 存在得分<70分的维度
  关键缺陷: 存在明显的质量问题
  风险隐患: 可能存在传播或品牌风险

特征表现:
  - 质量不达标: 未达到平台发布的基本质量要求
  - 用户价值不足: 为用户提供的价值有限
  - 传播风险存在: 可能存在传播效果不佳的风险
  - 商业价值微弱: 商业价值实现可能性较低

处理建议:
  - 暂缓发布: 建议暂缓发布，先进行优化
  - 重点改进: 针对问题维度进行重点改进
  - 质量提升: 全面提升内容质量水平
  - 风险规避: 识别和规避潜在风险

优化方向:
  - 内容重构: 可能需要对内容进行重新构思和重写
  - 角度调整: 调整内容角度和表达方式
  - 价值增强: 增强内容的用户价值和实用性
  - 风险消除: 识别并消除潜在的内容风险
```

## 📊 质量验证报告生成

### 🎯 综合质量评分算法
```python
def generate_comprehensive_quality_score(validation_results):
    """
    综合质量评分算法
    """
    # 权重分配
    weights = {
        "content_quality": 0.35,      # 内容质量权重35%
        "platform_adaptation": 0.25,  # 平台适配权重25%
        "user_experience": 0.25,      # 用户体验权重25%
        "effect_prediction": 0.15     # 效果预测权重15%
    }
    
    # 计算加权得分
    weighted_scores = {}
    for dimension, weight in weights.items():
        if dimension in validation_results:
            weighted_scores[dimension] = validation_results[dimension]["score"] * weight
    
    # 综合得分
    total_score = sum(weighted_scores.values())
    
    # 质量等级判定
    if total_score >= 90:
        quality_level = "优秀级"
        recommendation = "可重点推广的优质内容"
    elif total_score >= 75:
        quality_level = "良好级"
        recommendation = "符合发布标准的良好内容"
    else:
        quality_level = "待优化级"
        recommendation = "需要优化改进后再发布"
    
    return {
        "total_score": total_score,
        "weighted_scores": weighted_scores,
        "quality_level": quality_level,
        "recommendation": recommendation
    }
```

### 📋 问题诊断报告
```yaml
问题识别系统:
  内容问题:
    - 语言表达问题: 语法错误、表达不清、用词不当
    - 逻辑结构问题: 逻辑混乱、结构不清、论证不足
    - 价值密度问题: 价值不足、空洞无物、缺乏实用性
    - 原创性问题: 内容重复、观点陈旧、缺乏创新

  平台适配问题:
    - 风格不匹配: 不符合平台特色和用户偏好
    - 算法不友好: SEO效果差、推荐概率低
    - 用户不匹配: 目标用户定位不准确
    - 传播不适配: 分享传播机制设计不当

  用户体验问题:
    - 可读性差: 阅读难度过高、理解门槛过高
    - 吸引力不足: 标题平淡、内容无趣、缺乏亮点
    - 互动性弱: 互动设计不足、参与门槛过高
    - 实用性差: 缺乏实用价值、操作性不强

  效果预测问题:
    - 传播潜力弱: 缺乏传播亮点、分享动机不足
    - 转化预期低: 转化设计不当、转化路径不清
    - 商业价值小: 商业化潜力有限、变现路径不明
    - 风险隐患高: 存在内容风险、品牌风险
```

### 💡 优化建议生成
```yaml
优化建议框架:
  即时优化建议 (可立即执行):
    - 文字修改: 修正语法错误、优化用词表达
    - 结构调整: 优化段落结构、改善逻辑顺序
    - 标题优化: 提升标题吸引力、增强关键词效果
    - 互动增强: 增加互动元素、优化互动设计

  短期优化建议 (1-3天内完成):
    - 内容补充: 增加价值内容、补充关键信息
    - 案例更新: 更换更好的案例、增加新鲜素材
    - 视觉优化: 优化排版设计、增加视觉元素
    - 平台适配: 针对平台特色进行内容调整

  中期优化建议 (1-2周内完成):
    - 角度重构: 重新选择内容角度、调整表达方式
    - 价值升级: 提升内容价值层次、增强实用性
    - 创新元素: 增加创新元素、突出独特性
    - 效果设计: 优化传播效果、增强转化设计

  长期优化建议 (持续改进):
    - 模式总结: 总结成功模式、形成标准模板
    - 能力提升: 提升写作能力、增强专业水平
    - 创新探索: 探索新的创作形式、开拓新领域
    - 系统建设: 建设内容创作系统、形成规模效应
```

## 🎉 模块核心优势

### 🌟 验证全面准确
- **多维验证**: 从4个维度全面验证内容质量
- **标准精准**: 针对双平台特色制定精准验证标准
- **算法智能**: 采用智能算法提升验证准确性

### 🚀 建议实用有效
- **问题精准**: 精准识别内容存在的具体问题
- **建议具体**: 提供具体可操作的优化建议
- **效果可预期**: 优化建议的效果可预期可验证

### 💡 持续优化改进
- **反馈循环**: 建立持续的质量反馈优化循环
- **标准迭代**: 基于效果数据持续优化验证标准
- **能力提升**: 通过验证过程提升整体创作能力

---

*🔍 写作质量验证器 - 确保每篇内容都达到双平台的最高质量标准！* 🚀 