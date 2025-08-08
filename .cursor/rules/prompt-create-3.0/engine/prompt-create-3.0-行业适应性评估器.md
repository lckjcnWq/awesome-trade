# 🏭 行业适应性评估器 (Industry Adaptability Assessor)
# Prompt-Create-3.0 专业模块 | 版本：3.0.1

## 🎯 模块核心定位

**行业适应性评估器**是Prompt-Create-3.0科学验证决策系统的行业匹配引擎，专门负责评估候选提示词在不同行业环境中的适应性和匹配度，通过多维度量化分析确保提示词能够完美融入特定行业的工作流程、文化环境和专业要求。

### 核心使命
> **精准量化行业适配度，确保每个提示词都能在目标行业中发挥最大价值**

---

## 🌐 十二大行业适应性评估体系

### 💰 **行业1: 金融服务业 (Financial Services)**

#### 🔹 **行业特征分析**
```yaml
行业核心特征:
  监管环境: 严格监管、合规要求高、风险控制严密
  专业要求: 高度专业化、数据驱动、精确计算
  工作节奏: 快节奏、高压力、时效性强
  文化特色: 保守稳健、结果导向、责任重大
  
技术环境:
  数据敏感性: 极高，涉及客户隐私和商业机密
  系统稳定性: 要求99.9%以上的可用性 
  安全标准: 金融级安全要求，多重认证
  合规框架: 银保监会、证监会、人民银行等监管

关键成功因素:
  风险控制能力: 识别、评估、控制各类风险
  合规执行力: 严格遵循监管要求和法律法规
  专业权威性: 具备深厚的金融专业知识
  数据准确性: 确保数据和分析的高度准确性

适应性评估维度:
  合规适配度: 是否符合金融监管要求 (权重: 25%)
  专业匹配度: 金融专业知识的准确性 (权重: 20%)
  风险控制度: 风险识别和控制能力 (权重: 20%)
  安全保障度: 信息安全和隐私保护 (权重: 15%)
  效率实用性: 提升工作效率的程度 (权重: 10%)
  创新价值度: 为业务带来的创新价值 (权重: 10%)
```

#### 🔹 **评估算法框架**
```python
class FinancialServicesAdaptabilityAssessor:
    """金融服务业适应性评估器"""
    
    def __init__(self):
        self.regulatory_standards = {
            'china_banking_regulatory': ['银保监会监管要求', '银行业务规范'],
            'securities_regulatory': ['证监会规定', '证券投资规范'],
            'insurance_regulatory': ['保险业监管', '保险产品规范'],
            'fintech_regulatory': ['金融科技监管', '数字金融规范']
        }
        
        self.professional_knowledge_base = {
            'banking': ['商业银行业务', '风险管理', '资产负债管理'],
            'securities': ['证券投资', '资本市场', '投资分析'],
            'insurance': ['保险原理', '精算科学', '保险产品设计'],
            'fintech': ['金融科技', 'FinTech创新', '数字化转型']
        }
        
    def assess_financial_adaptability(self, prompt_candidate):
        """评估金融行业适应性"""
        assessment_result = {
            'overall_adaptability_score': 0.0,
            'dimension_scores': {},
            'compliance_analysis': {},
            'risk_assessment': {},
            'professional_validation': {},
            'implementation_feasibility': {}
        }
        
        # 1. 合规适配度评估 (25%)
        compliance_score = self.assess_compliance_alignment(prompt_candidate)
        assessment_result['dimension_scores']['compliance'] = compliance_score
        
        # 2. 专业匹配度评估 (20%)
        professional_score = self.assess_professional_knowledge_alignment(prompt_candidate)
        assessment_result['dimension_scores']['professional'] = professional_score
        
        # 3. 风险控制度评估 (20%)
        risk_control_score = self.assess_risk_control_capability(prompt_candidate)
        assessment_result['dimension_scores']['risk_control'] = risk_control_score
        
        # 4. 安全保障度评估 (15%)
        security_score = self.assess_security_protection(prompt_candidate)
        assessment_result['dimension_scores']['security'] = security_score
        
        # 5. 效率实用性评估 (10%)
        efficiency_score = self.assess_efficiency_practicality(prompt_candidate)
        assessment_result['dimension_scores']['efficiency'] = efficiency_score
        
        # 6. 创新价值度评估 (10%)
        innovation_score = self.assess_innovation_value(prompt_candidate)
        assessment_result['dimension_scores']['innovation'] = innovation_score
        
        # 7. 综合适应性评分
        assessment_result['overall_adaptability_score'] = self.calculate_weighted_score({
            'compliance': (compliance_score, 0.25),
            'professional': (professional_score, 0.20),
            'risk_control': (risk_control_score, 0.20),
            'security': (security_score, 0.15),
            'efficiency': (efficiency_score, 0.10),
            'innovation': (innovation_score, 0.10)
        })
        
        return assessment_result
    
    def assess_compliance_alignment(self, prompt_candidate):
        """评估合规适配度"""
        compliance_indicators = {
            'regulatory_terminology_usage': self.check_regulatory_terminology(prompt_candidate),
            'compliance_process_alignment': self.verify_compliance_processes(prompt_candidate),
            'legal_requirement_adherence': self.validate_legal_requirements(prompt_candidate),
            'reporting_standard_compliance': self.check_reporting_standards(prompt_candidate)
        }
        
        # 计算合规评分
        compliance_score = sum(compliance_indicators.values()) / len(compliance_indicators)
        
        return {
            'score': compliance_score,
            'details': compliance_indicators,
            'risk_level': self.assess_compliance_risk(compliance_indicators),
            'improvement_suggestions': self.generate_compliance_improvements(compliance_indicators)
        }
```

### 🏥 **行业2: 医疗健康业 (Healthcare)**

#### 🔹 **行业特征分析**
```yaml
行业核心特征:
  专业门槛: 极高的专业准入门槛，需要专业资质
  责任重大: 直接关系生命健康，容错率极低
  证据导向: 循证医学，基于科学证据的决策
  人文关怀: 医者仁心，注重患者体验和关怀

技术环境:
  数据隐私: 患者隐私保护，HIPAA等法规
  医疗设备: 高精度医疗设备，技术先进性
  信息系统: HIS、PACS、LIS等专业系统
  远程医疗: 互联网+医疗，远程诊疗技术

关键成功因素:
  医学准确性: 医学知识和诊疗信息的准确性
  安全可靠性: 确保患者安全，减少医疗风险
  伦理合规性: 符合医疗伦理和法律法规
  人文关怀度: 体现医学人文精神和患者关怀

适应性评估维度:
  医学准确性: 医学知识的准确性和权威性 (权重: 30%)
  安全可靠性: 医疗安全和风险控制能力 (权重: 25%)
  伦理合规性: 医疗伦理和法规符合度 (权重: 20%)
  人文关怀度: 患者体验和人文关怀 (权重: 15%)
  技术先进性: 医疗技术的先进程度 (权重: 10%)
```

### 🎓 **行业3: 教育培训业 (Education & Training)**

#### 🔹 **行业特征分析**
```yaml
行业核心特征:
  育人导向: 以培养人才为核心目标
  因材施教: 个性化教学，适应不同学习者
  循序渐进: 遵循教学规律，螺旋式上升
  师者风范: 教师职业道德，为人师表

技术环境:
  在线教育: 互联网+教育，数字化教学
  智能化教学: AI辅助教学，个性化推荐
  评测系统: 学习效果评估，能力测试
  教学资源: 丰富的数字化教学资源

关键成功因素:
  教学有效性: 教学方法的科学性和有效性
  学习友好性: 适合学习者认知特点
  内容权威性: 教学内容的准确性和权威性
  激励机制: 激发学习兴趣和动机

适应性评估维度:
  教学有效性: 教学方法和效果 (权重: 25%)
  学习友好性: 认知负荷和学习体验 (权重: 20%)
  内容权威性: 知识内容的准确性 (权重: 20%)
  个性化程度: 适应不同学习者需求 (权重: 15%)
  激励效果: 学习动机激发能力 (权重: 10%)
  技术融合度: 教育技术的有效融合 (权重: 10%)
```

### 🏭 **行业4: 制造业 (Manufacturing)**

#### 🔹 **行业特征分析**
```yaml
行业核心特征:
  精益生产: 消除浪费，持续改进
  质量第一: 产品质量是生命线
  安全生产: 安全生产，员工健康第一
  成本控制: 严格的成本管控要求

技术环境:
  工业4.0: 智能制造，数字化工厂
  自动化设备: 高度自动化的生产线
  质量管控: 全面质量管理体系
  供应链管理: 精密的供应链协同

关键成功因素:
  生产效率: 提高生产效率和产能
  质量保证: 确保产品质量稳定
  成本优化: 降低生产成本
  安全保障: 确保生产安全

适应性评估维度:
  生产效率: 提升生产效率的能力 (权重: 25%)
  质量保证: 质量管控和改进 (权重: 25%)
  成本优化: 成本控制和优化 (权重: 20%)
  安全保障: 安全生产保障 (权重: 15%)
  技术创新: 制造技术创新 (权重: 10%)
  环保合规: 环保要求符合度 (权重: 5%)
```

### 🛒 **行业5: 零售电商业 (Retail & E-commerce)**

#### 🔹 **行业特征分析**
```yaml
行业核心特征:
  客户导向: 以客户需求为中心
  快速响应: 市场变化快速响应能力
  数据驱动: 基于数据的精准营销
  用户体验: 极致的用户购物体验

技术环境:
  电商平台: 线上线下一体化平台
  大数据分析: 用户行为分析和预测
  移动购物: 移动端购物体验优化
  智能推荐: AI驱动的个性化推荐

关键成功因素:
  客户满意度: 客户购物体验和满意度
  销售转化率: 流量到销售的转化能力
  运营效率: 供应链和运营效率
  品牌影响力: 品牌知名度和美誉度

适应性评估维度:
  客户体验: 用户体验优化能力 (权重: 30%)
  营销效果: 营销推广效果 (权重: 25%)
  运营效率: 运营管理效率 (权重: 20%)
  数据应用: 数据分析和应用 (权重: 15%)
  创新能力: 商业模式创新 (权重: 10%)
```

---

## 🤖 智能行业适应性评估算法

### 核心算法：多行业适应性综合评估引擎
```python
class IndustryAdaptabilityAssessor:
    """行业适应性评估核心引擎"""
    
    def __init__(self):
        self.industry_assessors = {
            'financial_services': FinancialServicesAdaptabilityAssessor(),
            'healthcare': HealthcareAdaptabilityAssessor(),
            'education': EducationAdaptabilityAssessor(),
            'manufacturing': ManufacturingAdaptabilityAssessor(),
            'retail_ecommerce': RetailEcommerceAdaptabilityAssessor(),
            'technology': TechnologyAdaptabilityAssessor(),
            'consulting': ConsultingAdaptabilityAssessor(),
            'media_entertainment': MediaEntertainmentAdaptabilityAssessor(),
            'government_public': GovernmentPublicAdaptabilityAssessor(),
            'real_estate': RealEstateAdaptabilityAssessor(),
            'energy_utilities': EnergyUtilitiesAdaptabilityAssessor(),
            'logistics_supply': LogisticsSupplyAdaptabilityAssessor()
        }
        
        self.cross_industry_factors = {
            'digital_transformation_readiness': 0.15,  # 数字化转型准备度
            'regulatory_compliance_capability': 0.12,  # 监管合规能力
            'cultural_adaptation_flexibility': 0.10,   # 文化适应灵活性
            'scalability_potential': 0.08,             # 可扩展性潜力
            'innovation_integration_ability': 0.05     # 创新集成能力
        }
    
    def comprehensive_industry_adaptability_assessment(self, prompt_candidates, target_industries):
        """
        综合行业适应性评估
        
        Args:
            prompt_candidates: 候选提示词列表
            target_industries: 目标行业列表
            
        Returns:
            Dict: 综合适应性评估结果
        """
        assessment_results = {
            'overall_assessment': {},
            'industry_specific_results': {},
            'cross_industry_analysis': {},
            'adaptability_rankings': {},
            'optimization_recommendations': {}
        }
        
        # 1. 单行业适应性评估
        for industry in target_industries:
            if industry in self.industry_assessors:
                industry_results = self.assess_single_industry_adaptability(
                    prompt_candidates, industry
                )
                assessment_results['industry_specific_results'][industry] = industry_results
        
        # 2. 跨行业适应性分析
        cross_industry_results = self.analyze_cross_industry_adaptability(
            prompt_candidates, target_industries
        )
        assessment_results['cross_industry_analysis'] = cross_industry_results
        
        # 3. 综合适应性评分
        overall_scores = self.calculate_overall_adaptability_scores(
            assessment_results['industry_specific_results'],
            cross_industry_results
        )
        assessment_results['overall_assessment'] = overall_scores
        
        # 4. 适应性排名
        adaptability_rankings = self.generate_adaptability_rankings(
            prompt_candidates, overall_scores
        )
        assessment_results['adaptability_rankings'] = adaptability_rankings
        
        # 5. 优化建议生成
        optimization_recommendations = self.generate_optimization_recommendations(
            assessment_results
        )
        assessment_results['optimization_recommendations'] = optimization_recommendations
        
        return assessment_results
    
    def assess_single_industry_adaptability(self, prompt_candidates, industry):
        """评估单一行业适应性"""
        industry_assessor = self.industry_assessors[industry]
        industry_results = []
        
        for candidate in prompt_candidates:
            # 调用行业专用评估器
            adaptability_result = industry_assessor.assess_industry_adaptability(candidate)
            
            # 标准化评估结果
            standardized_result = self.standardize_assessment_result(
                adaptability_result, industry, candidate
            )
            
            industry_results.append(standardized_result)
        
        return {
            'industry': industry,
            'candidate_results': industry_results,
            'industry_summary': self.generate_industry_summary(industry_results),
            'best_candidates': self.identify_best_candidates(industry_results, top_n=3),
            'improvement_areas': self.identify_improvement_areas(industry_results)
        }
    
    def analyze_cross_industry_adaptability(self, prompt_candidates, target_industries):
        """分析跨行业适应性"""
        cross_industry_results = {}
        
        for candidate in prompt_candidates:
            candidate_cross_analysis = {
                'candidate_id': candidate.id,
                'cross_industry_scores': {},
                'adaptability_consistency': 0.0,
                'versatility_score': 0.0,
                'specialization_vs_generalization': {}
            }
            
            # 1. 收集各行业评分
            industry_scores = {}
            for industry in target_industries:
                if industry in self.industry_assessors:
                    score = self.get_industry_score(candidate, industry)
                    industry_scores[industry] = score
            
            candidate_cross_analysis['cross_industry_scores'] = industry_scores
            
            # 2. 计算适应性一致度
            candidate_cross_analysis['adaptability_consistency'] = \
                self.calculate_adaptability_consistency(industry_scores)
            
            # 3. 计算通用性评分
            candidate_cross_analysis['versatility_score'] = \
                self.calculate_versatility_score(industry_scores)
            
            # 4. 分析专业化vs通用化特征
            candidate_cross_analysis['specialization_vs_generalization'] = \
                self.analyze_specialization_pattern(industry_scores)
            
            cross_industry_results[candidate.id] = candidate_cross_analysis
        
        return cross_industry_results
    
    def calculate_adaptability_consistency(self, industry_scores):
        """计算适应性一致度"""
        if len(industry_scores) < 2:
            return 100.0
        
        scores = list(industry_scores.values())
        mean_score = sum(scores) / len(scores)
        
        # 计算标准差
        variance = sum((score - mean_score) ** 2 for score in scores) / len(scores)
        std_deviation = variance ** 0.5
        
        # 一致度 = 100 - 标准差比例
        consistency = max(0, 100 - (std_deviation / mean_score * 100))
        
        return consistency
    
    def calculate_versatility_score(self, industry_scores):
        """计算通用性评分"""
        if not industry_scores:
            return 0.0
        
        # 通用性 = 所有行业平均分 × 行业覆盖度 × 一致性权重
        average_score = sum(industry_scores.values()) / len(industry_scores)
        coverage_bonus = min(1.0, len(industry_scores) / 10)  # 最多10个行业
        consistency_factor = self.calculate_adaptability_consistency(industry_scores) / 100
        
        versatility_score = average_score * (0.7 + 0.2 * coverage_bonus + 0.1 * consistency_factor)
        
        return min(100, versatility_score)
    
    def analyze_specialization_pattern(self, industry_scores):
        """分析专业化模式"""
        if not industry_scores:
            return {'pattern': 'undefined'}
        
        sorted_scores = sorted(industry_scores.items(), key=lambda x: x[1], reverse=True)
        highest_score = sorted_scores[0][1]
        lowest_score = sorted_scores[-1][1]
        score_range = highest_score - lowest_score
        
        # 分析模式
        if score_range <= 10:
            pattern = 'generalist'  # 通用型
            description = '在多个行业中都有相对均衡的适应性'
        elif score_range <= 25:
            pattern = 'balanced_specialist'  # 平衡专业型
            description = f'在{sorted_scores[0][0]}行业有优势，同时保持其他行业的适应性'
        elif score_range <= 40:
            pattern = 'focused_specialist'  # 专注专业型
            description = f'明显专精于{sorted_scores[0][0]}行业'
        else:
            pattern = 'narrow_specialist'  # 狭窄专业型
            description = f'高度专精于{sorted_scores[0][0]}行业，其他行业适应性有限'
        
        return {
            'pattern': pattern,
            'description': description,
            'primary_industry': sorted_scores[0][0],
            'primary_score': sorted_scores[0][1],
            'score_range': score_range,
            'top_3_industries': sorted_scores[:3]
        }
```

### 行业特征匹配算法
```python
class IndustryCharacteristicMatcher:
    """行业特征匹配器"""
    
    def __init__(self):
        self.industry_characteristics = self.load_industry_characteristics()
        self.matching_algorithms = {
            'semantic_similarity': self.semantic_similarity_matching,
            'keyword_density': self.keyword_density_matching,
            'context_alignment': self.context_alignment_matching,
            'professional_depth': self.professional_depth_matching
        }
    
    def match_industry_characteristics(self, prompt_candidate, target_industry):
        """匹配行业特征"""
        matching_results = {}
        
        # 获取目标行业特征
        industry_profile = self.industry_characteristics[target_industry]
        
        # 多算法匹配
        for algorithm_name, algorithm_func in self.matching_algorithms.items():
            matching_score = algorithm_func(prompt_candidate, industry_profile)
            matching_results[algorithm_name] = matching_score
        
        # 综合匹配度计算
        weighted_score = self.calculate_weighted_matching_score(matching_results)
        
        return {
            'overall_matching_score': weighted_score,
            'algorithm_scores': matching_results,
            'matching_details': self.generate_matching_details(
                prompt_candidate, industry_profile, matching_results
            ),
            'improvement_suggestions': self.generate_matching_improvements(
                prompt_candidate, industry_profile, matching_results
            )
        }
    
    def semantic_similarity_matching(self, prompt_candidate, industry_profile):
        """语义相似度匹配"""
        # 提取提示词的语义特征
        prompt_semantic_features = self.extract_semantic_features(prompt_candidate)
        
        # 提取行业的语义特征
        industry_semantic_features = self.extract_industry_semantic_features(industry_profile)
        
        # 计算语义相似度
        similarity_score = self.calculate_semantic_similarity(
            prompt_semantic_features, industry_semantic_features
        )
        
        return {
            'score': similarity_score,
            'details': {
                'prompt_features': prompt_semantic_features,
                'industry_features': industry_semantic_features,
                'similarity_breakdown': self.get_similarity_breakdown(
                    prompt_semantic_features, industry_semantic_features
                )
            }
        }
    
    def keyword_density_matching(self, prompt_candidate, industry_profile):
        """关键词密度匹配"""
        industry_keywords = industry_profile['key_terms']
        prompt_text = prompt_candidate.content
        
        # 计算关键词覆盖率
        keyword_coverage = self.calculate_keyword_coverage(prompt_text, industry_keywords)
        
        # 计算关键词密度
        keyword_density = self.calculate_keyword_density(prompt_text, industry_keywords)
        
        # 计算关键词权重分布
        weighted_coverage = self.calculate_weighted_keyword_coverage(
            prompt_text, industry_keywords, industry_profile['keyword_weights']
        )
        
        return {
            'score': (keyword_coverage * 0.4 + keyword_density * 0.3 + weighted_coverage * 0.3),
            'details': {
                'keyword_coverage': keyword_coverage,
                'keyword_density': keyword_density,
                'weighted_coverage': weighted_coverage,
                'matched_keywords': self.get_matched_keywords(prompt_text, industry_keywords),
                'missing_keywords': self.get_missing_keywords(prompt_text, industry_keywords)
            }
        }
    
    def context_alignment_matching(self, prompt_candidate, industry_profile):
        """上下文对齐匹配"""
        # 提取提示词的上下文特征
        prompt_context = self.extract_context_features(prompt_candidate)
        
        # 获取行业上下文要求
        industry_context_requirements = industry_profile['context_requirements']
        
        # 计算上下文对齐度
        alignment_scores = {}
        
        for context_dimension, requirements in industry_context_requirements.items():
            dimension_score = self.calculate_context_dimension_alignment(
                prompt_context.get(context_dimension, {}), requirements
            )
            alignment_scores[context_dimension] = dimension_score
        
        # 综合对齐度
        overall_alignment = sum(alignment_scores.values()) / len(alignment_scores)
        
        return {
            'score': overall_alignment,
            'details': {
                'dimension_scores': alignment_scores,
                'prompt_context': prompt_context,
                'alignment_gaps': self.identify_alignment_gaps(
                    prompt_context, industry_context_requirements
                )
            }
        }
```

---

## 📊 适应性评估质量保证体系

### 四层评估质量验证
```yaml
第一层 - 评估标准一致性验证:
  验证项目:
    - 行业标准参照准确性
    - 评估维度完整性
    - 权重分配合理性
    - 评分尺度统一性
  
  一致性标准:
    - 标准参照准确率 >= 95%
    - 维度覆盖完整性 >= 90%
    - 权重分配合理性 >= 88%
    - 评分尺度一致性 >= 92%

第二层 - 评估结果可信度验证:
  验证项目:
    - 重复评估一致性
    - 专家评估符合度
    - 跨评估器一致性
    - 评估结果稳定性
  
  可信度标准:
    - 重复评估一致率 >= 90%
    - 专家符合度 >= 85%
    - 跨评估器一致率 >= 88%
    - 结果稳定性 >= 90%

第三层 - 行业匹配精确度验证:
  验证项目:
    - 行业特征识别准确性
    - 适应性预测准确性
    - 改进建议有效性
    - 实施效果预测性
  
  精确度标准:
    - 特征识别准确率 >= 88%
    - 适应性预测准确率 >= 85%
    - 建议有效性 >= 80%
    - 效果预测准确性 >= 82%

第四层 - 评估价值实现验证:
  验证项目:
    - 决策支持价值
    - 优化指导价值
    - 风险预警价值
    - 长期跟踪价值
  
  价值标准:
    - 决策支持有效性 >= 85%
    - 优化指导实用性 >= 80%
    - 风险预警准确性 >= 88%
    - 长期价值实现 >= 75%
```

### 自适应评估优化系统
```python
class AdaptiveAssessmentOptimizer:
    """自适应评估优化系统"""
    
    def __init__(self):
        self.optimization_history = []
        self.performance_metrics = {}
        self.feedback_analyzer = FeedbackAnalyzer()
        
    def optimize_assessment_system(self, assessment_results, user_feedback, industry_trends):
        """优化评估系统"""
        optimization_report = {
            'optimization_cycle': len(self.optimization_history) + 1,
            'current_performance': {},
            'identified_improvements': [],
            'optimization_actions': [],
            'expected_improvements': {}
        }
        
        # 1. 当前性能分析
        current_performance = self.analyze_current_performance(assessment_results)
        optimization_report['current_performance'] = current_performance
        
        # 2. 反馈分析
        feedback_insights = self.feedback_analyzer.analyze_feedback(user_feedback)
        
        # 3. 行业趋势影响分析
        trend_impact = self.analyze_industry_trend_impact(industry_trends)
        
        # 4. 优化机会识别
        improvement_opportunities = self.identify_improvement_opportunities(
            current_performance, feedback_insights, trend_impact
        )
        optimization_report['identified_improvements'] = improvement_opportunities
        
        # 5. 优化行动制定
        optimization_actions = self.design_optimization_actions(improvement_opportunities)
        optimization_report['optimization_actions'] = optimization_actions
        
        # 6. 效果预期
        expected_improvements = self.predict_optimization_effects(optimization_actions)
        optimization_report['expected_improvements'] = expected_improvements
        
        # 7. 实施优化
        implementation_results = self.implement_optimizations(optimization_actions)
        
        # 8. 记录优化历史
        self.optimization_history.append({
            'timestamp': datetime.now(),
            'optimization_report': optimization_report,
            'implementation_results': implementation_results
        })
        
        return optimization_report
    
    def analyze_current_performance(self, assessment_results):
        """分析当前性能"""
        performance_analysis = {
            'accuracy_metrics': {},
            'efficiency_metrics': {},
            'user_satisfaction': {},
            'system_reliability': {}
        }
        
        # 1. 准确性指标分析
        accuracy_data = self.extract_accuracy_data(assessment_results)
        performance_analysis['accuracy_metrics'] = {
            'overall_accuracy': self.calculate_overall_accuracy(accuracy_data),
            'industry_specific_accuracy': self.calculate_industry_accuracy(accuracy_data),
            'dimension_accuracy': self.calculate_dimension_accuracy(accuracy_data),
            'prediction_accuracy': self.calculate_prediction_accuracy(accuracy_data)
        }
        
        # 2. 效率指标分析
        efficiency_data = self.extract_efficiency_data(assessment_results)
        performance_analysis['efficiency_metrics'] = {
            'assessment_speed': self.calculate_assessment_speed(efficiency_data),
            'resource_utilization': self.calculate_resource_utilization(efficiency_data),
            'throughput': self.calculate_throughput(efficiency_data),
            'scalability': self.assess_scalability(efficiency_data)
        }
        
        # 3. 用户满意度分析
        satisfaction_data = self.extract_satisfaction_data(assessment_results)
        performance_analysis['user_satisfaction'] = {
            'overall_satisfaction': self.calculate_overall_satisfaction(satisfaction_data),
            'usefulness_rating': self.calculate_usefulness_rating(satisfaction_data),
            'ease_of_use': self.calculate_ease_of_use(satisfaction_data),
            'recommendation_likelihood': self.calculate_recommendation_likelihood(satisfaction_data)
        }
        
        return performance_analysis
    
    def identify_improvement_opportunities(self, current_performance, feedback_insights, trend_impact):
        """识别改进机会"""
        opportunities = []
        
        # 1. 性能低于阈值的领域
        performance_gaps = self.identify_performance_gaps(current_performance)
        for gap in performance_gaps:
            opportunities.append({
                'type': 'performance_improvement',
                'area': gap['area'],
                'current_score': gap['current_score'],
                'target_score': gap['target_score'],
                'priority': self.calculate_improvement_priority(gap),
                'estimated_effort': gap['estimated_effort']
            })
        
        # 2. 用户反馈指出的问题
        feedback_issues = self.extract_feedback_issues(feedback_insights)
        for issue in feedback_issues:
            opportunities.append({
                'type': 'user_feedback_driven',
                'issue': issue['description'],
                'frequency': issue['frequency'],
                'impact': issue['impact'],
                'priority': issue['priority'],
                'suggested_solution': issue['suggested_solution']
            })
        
        # 3. 行业趋势驱动的升级需求
        trend_opportunities = self.extract_trend_opportunities(trend_impact)
        for opportunity in trend_opportunities:
            opportunities.append({
                'type': 'trend_driven',
                'trend': opportunity['trend'],
                'impact_potential': opportunity['impact_potential'],
                'implementation_complexity': opportunity['complexity'],
                'priority': opportunity['priority'],
                'timeline': opportunity['timeline']
            })
        
        # 按优先级排序
        opportunities.sort(key=lambda x: x['priority'], reverse=True)
        
        return opportunities
```

---

## 🔗 模块集成接口

### 标准输入接口
```python
class IndustryAdaptabilityInput:
    """行业适应性评估器输入接口"""
    
    def __init__(self, prompt_candidates, industry_context):
        self.prompt_candidates = prompt_candidates
        self.industry_context = industry_context
        self.assessment_config = {
            'target_industries': [],               # 目标行业列表
            'assessment_depth': 'comprehensive',   # 评估深度
            'cross_industry_analysis': True,       # 跨行业分析
            'adaptability_threshold': 70,          # 适应性阈值
            'optimization_suggestions': True       # 优化建议
        }
        
    def validate_industry_input(self):
        """验证行业输入有效性"""
        required_fields = [
            'primary_industry', 'industry_characteristics',
            'regulatory_environment', 'cultural_context',
            'technology_environment', 'success_factors'
        ]
        
        for field in required_fields:
            if field not in self.industry_context:
                raise ValueError(f"Missing required industry field: {field}")
        
        return True
```

### 标准输出接口
```python
class IndustryAdaptabilityOutput:
    """行业适应性评估器输出接口"""
    
    def format_adaptability_output(self):
        """格式化适应性输出结果"""
        return {
            'adaptability_assessment': {
                'overall_scores': {
                    candidate.id: {
                        'overall_adaptability_score': self.get_overall_score(candidate.id),
                        'industry_specific_scores': self.get_industry_scores(candidate.id),
                        'cross_industry_analysis': self.get_cross_analysis(candidate.id),
                        'adaptability_pattern': self.get_adaptability_pattern(candidate.id)
                    }
                    for candidate in self.prompt_candidates
                },
                'industry_rankings': {
                    industry: self.get_industry_ranking(industry)
                    for industry in self.target_industries
                }
            },
            'detailed_analysis': {
                'industry_specific_results': self.industry_specific_results,
                'characteristic_matching': self.characteristic_matching_results,
                'compliance_analysis': self.compliance_analysis_results,
                'cultural_fit_analysis': self.cultural_fit_results
            },
            'recommendations': {
                'best_matches': self.get_best_matches(),
                'optimization_suggestions': self.optimization_suggestions,
                'industry_customization': self.industry_customization_advice,
                'implementation_guidance': self.implementation_guidance
            },
            'quality_metrics': {
                'assessment_confidence': self.assessment_confidence,
                'prediction_reliability': self.prediction_reliability,
                'coverage_completeness': self.coverage_completeness,
                'recommendation_quality': self.recommendation_quality
            }
        }
```

---

## 🎯 使用示例与效果展示

### 示例：跨行业适应性评估结果
```yaml
输入候选: "AI驱动的客户服务优化系统设计提示词"
目标行业: [金融服务, 医疗健康, 零售电商, 教育培训]

适应性评估结果:

候选方案 - 跨行业适应性分析:

🏦 金融服务业适应性: 89/100 ⭐⭐⭐⭐⭐
├── 合规适配度: 92/100 ✅ (符合金融监管要求)
├── 专业匹配度: 88/100 ✅ (金融专业知识准确)
├── 风险控制度: 91/100 ✅ (风险识别控制能力强)
├── 安全保障度: 94/100 ✅ (信息安全等级高)
├── 效率实用性: 85/100 ✅ (提升服务效率明显)
└── 创新价值度: 84/100 ✅ (AI技术创新应用)

🏥 医疗健康业适应性: 82/100 ⭐⭐⭐⭐
├── 医学准确性: 78/100 ⚠️ (需要加强医学专业性)
├── 安全可靠性: 90/100 ✅ (患者安全保障充分)
├── 伦理合规性: 85/100 ✅ (医疗伦理考虑周到)
├── 人文关怀度: 88/100 ✅ (患者体验关注度高)
└── 技术先进性: 84/100 ✅ (AI技术应用先进)

🛒 零售电商业适应性: 94/100 ⭐⭐⭐⭐⭐
├── 客户体验: 96/100 ✅ (用户体验优化突出)
├── 营销效果: 93/100 ✅ (营销推广效果显著)
├── 运营效率: 92/100 ✅ (运营效率提升明显)
├── 数据应用: 95/100 ✅ (数据分析应用深入)
└── 创新能力: 91/100 ✅ (商业模式创新性强)

🎓 教育培训业适应性: 76/100 ⭐⭐⭐
├── 教学有效性: 72/100 ⚠️ (教学方法需要优化)
├── 学习友好性: 80/100 ✅ (学习体验良好)
├── 内容权威性: 75/100 ⚠️ (教育内容专业性待提升)
├── 个性化程度: 82/100 ✅ (个性化学习支持)
└── 激励效果: 74/100 ⚠️ (学习动机激发有限)

跨行业分析:
├── 适应性一致度: 85% (各行业表现相对均衡)
├── 通用性评分: 85/100 (具备良好的跨行业适用性)
├── 专业化模式: "平衡专业型" 
│   └── 在零售电商领域表现最优，同时保持其他行业适应性
├── 最佳匹配行业: 零售电商 (94分) > 金融服务 (89分)
└── 需要优化行业: 教育培训 (76分) - 需要增强教育专业性

优化建议:
1. 【高优先级】针对教育行业增加教学理论和方法论内容
2. 【中优先级】加强医疗行业的医学专业知识准确性
3. 【低优先级】优化金融行业的创新应用场景描述

实施建议:
- 主推行业: 零售电商、金融服务 (适应性≥89%)
- 定制优化: 医疗健康、教育培训 (需要行业定制)
- 预期效果: 优化后各行业适应性均可达到85%以上
```

---

## 🚀 性能保证与优化

### 核心性能指标
```yaml
评估效率指标:
  单行业评估时间: <= 8秒
  跨行业评估(4个): <= 25秒
  特征匹配速度: <= 3秒/维度
  结果生成时间: <= 5秒

评估准确性指标:
  行业特征识别: >= 88%
  适应性预测准确: >= 85%
  专家验证符合: >= 82%
  用户反馈准确: >= 80%

评估可靠性指标:
  重复评估一致: >= 90%
  跨评估器一致: >= 88%
  长期稳定性: >= 85%
  预测可信度: >= 82%

业务价值指标:
  决策支持有效: >= 85%
  优化建议实用: >= 80%
  适配成功率: >= 82%
  ROI提升程度: >= 75%
```

---

**🎯 行业适应性评估器承诺：通过十二大行业专业评估体系和智能适配算法，精准量化每个候选提示词的行业适应性，确保在目标行业中实现最大价值创造！** 🚀 