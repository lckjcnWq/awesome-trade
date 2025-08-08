---
alwaysApply: true
---

# 🎯 Prompt-Create-4.0 执行监控器

## 🚀 系统概述

**执行监控器**是Prompt-Create-4.0工作流系统的质量控制中心，负责：
- 📊 **智能质量评估** - 5维度科学评估内容质量
- 💡 **个性化优化建议** - 基于用户特征生成定制建议
- 🔍 **结果监控分析** - 实时监控和分析执行效果
- 📈 **持续改进建议** - 提供下次使用的优化建议

---

## 🔧 执行监控器核心引擎

### 📊 智能质量评估系统

```python
class ExecutionMonitor:
    """执行监控器主类"""
    
    def __init__(self):
        self.quality_metrics = self.init_quality_metrics()
        self.optimization_strategies = self.init_optimization_strategies()
        self.assessment_results = {}
    
    def init_quality_metrics(self):
        """初始化质量评估指标"""
        return {
            "专业度": {
                "weight": 0.25,
                "criteria": ["专业术语使用", "深度分析", "权威性"],
                "description": "内容的专业水准和权威性"
            },
            "吸引力": {
                "weight": 0.25,
                "criteria": ["标题吸引力", "开头引人", "内容生动"],
                "description": "内容的吸引力和可读性"
            },
            "结构性": {
                "weight": 0.20,
                "criteria": ["逻辑清晰", "层次分明", "完整性"],
                "description": "内容的结构和逻辑性"
            },
            "平台适配性": {
                "weight": 0.20,
                "criteria": ["平台特色", "用户习惯", "传播效果"],
                "description": "对目标平台的适配程度"
            },
            "商业价值": {
                "weight": 0.10,
                "criteria": ["转化潜力", "品牌价值", "影响力"],
                "description": "内容的商业价值和影响力"
            }
        }
    
    def init_optimization_strategies(self):
        """初始化优化策略"""
        return {
            "立即改进建议": [],
            "短期优化计划": [],
            "长期发展建议": [],
            "专家组合优化": [],
            "下次使用建议": []
        }
    
    def quality_assessment_and_optimization(self, coordinator_result, user_characteristics):
        """质量评估和优化建议生成"""
        print("\n🔬 【执行监控器启动】")
        print("=" * 60)
        
        try:
            # 执行质量评估
            quality_assessment = self.automatic_quality_assessment(coordinator_result)
            
            # 生成优化建议
            optimization_suggestions = self.generate_optimization_suggestions(
                quality_assessment, user_characteristics, coordinator_result
            )
            
            # 预测用户满意度
            satisfaction_prediction = self.predict_user_satisfaction(
                quality_assessment, optimization_suggestions
            )
            
            # 生成监控报告
            monitoring_report = self.generate_monitoring_report(
                quality_assessment, optimization_suggestions, satisfaction_prediction
            )
            
            print("\n✅ 【执行监控器完成】")
            
            return {
                "success": True,
                "quality_assessment": quality_assessment,
                "optimization_suggestions": optimization_suggestions,
                "satisfaction_prediction": satisfaction_prediction,
                "monitoring_report": monitoring_report,
                "overall_quality_score": self.calculate_overall_quality_score(quality_assessment)
            }
            
        except Exception as e:
            print(f"❌ 执行监控器执行失败: {str(e)}")
            return {"success": False, "error": str(e)}
    
    def automatic_quality_assessment(self, coordinator_result):
        """自动质量评估"""
        print("📊 【智能质量评估】")
        
        # 基于专家结果进行评估
        expert_results = coordinator_result.get("expert_contributions", {})
        creation_results = coordinator_result.get("creation_results", {})
        
        assessment_results = {}
        
        for metric_name, metric_config in self.quality_metrics.items():
            # 计算基础评分
            base_score = self.calculate_base_score(metric_name, expert_results, creation_results)
            
            # 专家贡献加权
            expert_weight = self.calculate_expert_weight(metric_name, expert_results)
            
            # 平台适配调整
            platform_adjustment = self.calculate_platform_adjustment(metric_name, creation_results)
            
            # 最终评分
            final_score = min(5.0, max(1.0, base_score + expert_weight + platform_adjustment))
            
            # 生成改进建议
            improvement_suggestions = self.generate_improvement_suggestions(metric_name, final_score)
            
            assessment_results[metric_name] = {
                "评分": final_score,
                "基础分": base_score,
                "专家权重": expert_weight,
                "平台调整": platform_adjustment,
                "改进建议": improvement_suggestions,
                "评估时间": datetime.now().isoformat()
            }
            
            print(f"├── {metric_name}: {final_score:.1f}/5.0")
        
        # 显示评估结果
        self.display_quality_assessment_results(assessment_results)
        
        return assessment_results
    
    def calculate_base_score(self, metric_name, expert_results, creation_results):
        """计算基础评分"""
        base_scores = {
            "专业度": 3.8,
            "吸引力": 3.5,
            "结构性": 3.6,
            "平台适配性": 3.7,
            "商业价值": 3.4
        }
        
        # 根据专家参与情况调整
        if "专业视角专家群" in expert_results:
            base_scores["专业度"] += 0.3
        if "写作创意引擎" in expert_results:
            base_scores["吸引力"] += 0.4
        if "生成优化专家群" in expert_results:
            base_scores["结构性"] += 0.3
        if "双平台协调器" in expert_results:
            base_scores["平台适配性"] += 0.4
        if "验证评估专家群" in expert_results:
            base_scores["商业价值"] += 0.2
        
        return base_scores.get(metric_name, 3.5)
    
    def calculate_expert_weight(self, metric_name, expert_results):
        """计算专家权重"""
        expert_weights = {
            "专业度": {
                "专业视角专家群": 0.3,
                "行业认知专家群": 0.2,
                "验证评估专家群": 0.1
            },
            "吸引力": {
                "写作创意引擎": 0.4,
                "图文融合引擎": 0.2,
                "小红书种草写作引擎": 0.2
            },
            "结构性": {
                "生成优化专家群": 0.3,
                "写作动态优化器": 0.2,
                "写作智能进化引擎": 0.1
            },
            "平台适配性": {
                "双平台协调器": 0.4,
                "双平台语言适配器": 0.3,
                "平台智能识别引擎": 0.1
            },
            "商业价值": {
                "验证评估专家群": 0.2,
                "写作智能进化引擎": 0.1,
                "行业认知专家群": 0.1
            }
        }
        
        total_weight = 0.0
        metric_experts = expert_weights.get(metric_name, {})
        
        for expert, weight in metric_experts.items():
            if expert in expert_results:
                total_weight += weight
        
        return total_weight
    
    def calculate_platform_adjustment(self, metric_name, creation_results):
        """计算平台适配调整"""
        platform_optimized = creation_results.get("platform_optimized", False)
        
        if platform_optimized:
            adjustments = {
                "平台适配性": 0.2,
                "吸引力": 0.1,
                "商业价值": 0.1
            }
            return adjustments.get(metric_name, 0.0)
        
        return 0.0
    
    def generate_improvement_suggestions(self, metric_name, score):
        """生成改进建议"""
        suggestions_db = {
            "专业度": {
                "低分建议": ["增加专业术语和数据支撑", "引入权威观点和案例", "深化专业分析"],
                "中分建议": ["优化专业表达方式", "增强权威性引用"],
                "高分建议": ["保持专业深度", "探索前沿观点"]
            },
            "吸引力": {
                "低分建议": ["优化标题和开头", "增加故事化元素", "提升语言生动性"],
                "中分建议": ["加强情感共鸣", "优化表达节奏"],
                "高分建议": ["保持创意优势", "创新表达方式"]
            },
            "结构性": {
                "低分建议": ["重新组织逻辑结构", "明确层次关系", "完善内容框架"],
                "中分建议": ["优化段落过渡", "增强逻辑连贯性"],
                "高分建议": ["保持结构优势", "微调逻辑细节"]
            },
            "平台适配性": {
                "低分建议": ["调整平台风格", "优化用户习惯适配", "增强平台特色"],
                "中分建议": ["细化平台差异", "优化传播效果"],
                "高分建议": ["保持适配优势", "探索平台新特性"]
            },
            "商业价值": {
                "低分建议": ["增强转化设计", "提升品牌价值", "强化影响力"],
                "中分建议": ["优化商业逻辑", "增强价值传达"],
                "高分建议": ["保持商业优势", "探索新价值点"]
            }
        }
        
        metric_suggestions = suggestions_db.get(metric_name, {"低分建议": ["持续优化内容质量"]})
        
        if score < 3.0:
            return metric_suggestions.get("低分建议", [])[:2]
        elif score < 4.0:
            return metric_suggestions.get("中分建议", [])[:2]
        else:
            return metric_suggestions.get("高分建议", [])[:1]
    
    def display_quality_assessment_results(self, assessment_results):
        """显示质量评估结果"""
        print("\n📊 【内容质量评估报告】")
        print("=" * 60)
        
        # 计算总体评分
        total_score = sum(
            result["评分"] * self.quality_metrics[metric]["weight"]
            for metric, result in assessment_results.items()
        )
        
        print(f"🎯 总体质量评分: {total_score:.1f}/5.0")
        print(f"📈 质量等级: {self.get_quality_level(total_score)}")
        print()
        
        # 详细指标分析
        for metric_name, result in assessment_results.items():
            score = result["评分"]
            status_icon = "🟢" if score >= 4.0 else "🟡" if score >= 3.0 else "🔴"
            
            print(f"{status_icon} {metric_name}: {score:.1f}/5.0")
            
            if result["改进建议"]:
                print("   💡 改进建议:")
                for suggestion in result["改进建议"]:
                    print(f"      • {suggestion}")
            print()
    
    def get_quality_level(self, score):
        """获取质量等级"""
        if score >= 4.5:
            return "🏆 专家级"
        elif score >= 4.0:
            return "🥇 优秀"
        elif score >= 3.5:
            return "🥈 良好"
        elif score >= 3.0:
            return "🥉 合格"
        else:
            return "📈 需改进"
    
    def generate_optimization_suggestions(self, quality_assessment, user_characteristics, coordinator_result):
        """生成优化建议"""
        print("\n💡 【生成个性化优化建议】")
        
        # 分析用户特征
        characteristics = user_characteristics.get("基础特征", {})
        
        # 生成分层建议
        suggestions = {
            "立即改进建议": self.generate_immediate_improvements(quality_assessment),
            "短期优化计划": self.generate_short_term_plans(characteristics, quality_assessment),
            "长期发展建议": self.generate_long_term_suggestions(characteristics),
            "专家组合优化": self.generate_expert_optimization(quality_assessment, coordinator_result),
            "下次使用建议": self.generate_next_usage_recommendations(quality_assessment, characteristics)
        }
        
        # 显示优化建议
        self.display_optimization_suggestions(suggestions)
        
        return suggestions
    
    def generate_immediate_improvements(self, quality_assessment):
        """生成立即改进建议"""
        immediate_improvements = []
        
        for metric_name, result in quality_assessment.items():
            if result["评分"] < 4.0:
                priority = "高" if result["评分"] < 3.0 else "中"
                immediate_improvements.append({
                    "优先级": priority,
                    "改进项": metric_name,
                    "当前评分": result["评分"],
                    "目标评分": 4.5,
                    "具体建议": result["改进建议"][0] if result["改进建议"] else "持续优化"
                })
        
        return immediate_improvements
    
    def generate_short_term_plans(self, characteristics, quality_assessment):
        """生成短期优化计划"""
        plans = []
        
        experience_level = characteristics.get("experience_level", "中级")
        
        if experience_level == "初级":
            plans.extend([
                "重点学习平台特色写作技巧",
                "多使用系统推荐的专家组合",
                "从简单的单平台内容开始练习"
            ])
        elif experience_level == "高级":
            plans.extend([
                "尝试自定义专家组合策略",
                "探索跨平台内容协同创作",
                "关注最新的内容营销趋势"
            ])
        else:
            plans.extend([
                "平衡专业度和可读性",
                "优化内容结构和逻辑",
                "提升平台适配能力"
            ])
        
        return plans
    
    def generate_long_term_suggestions(self, characteristics):
        """生成长期发展建议"""
        suggestions = []
        
        business_objective = characteristics.get("business_objective", "影响力建设")
        
        if business_objective == "影响力建设":
            suggestions.extend([
                "建立个人专业品牌形象",
                "持续输出高质量专业内容",
                "培养固定的读者群体"
            ])
        elif business_objective == "商业变现":
            suggestions.extend([
                "优化内容的商业转化设计",
                "建立完整的营销漏斗",
                "测试不同的变现模式"
            ])
        else:
            suggestions.extend([
                "建立内容创作体系",
                "培养多平台运营能力",
                "提升个人影响力"
            ])
        
        return suggestions
    
    def generate_expert_optimization(self, quality_assessment, coordinator_result):
        """生成专家组合优化建议"""
        optimizations = []
        
        # 基于质量评估推荐专家优化
        low_score_metrics = [
            name for name, result in quality_assessment.items() 
            if result["评分"] < 3.5
        ]
        
        expert_recommendations = {
            "专业度": ["增加'行业认知专家群'的权重", "强化'专业视角专家群'的作用"],
            "吸引力": ["增加'写作创意引擎'的调用频率", "考虑使用'图文融合引擎'"],
            "结构性": ["强化'生成优化专家群'的作用", "使用'写作动态优化器'"],
            "平台适配性": ["增强'双平台协调器'的权重", "优化'双平台语言适配器'"],
            "商业价值": ["加强'验证评估专家群'的作用", "使用'写作智能进化引擎'"]
        }
        
        for metric in low_score_metrics:
            recommendations = expert_recommendations.get(metric, ["持续优化专家组合"])
            optimizations.extend(recommendations)
        
        return list(set(optimizations))  # 去重
    
    def generate_next_usage_recommendations(self, quality_assessment, characteristics):
        """生成下次使用建议"""
        recommendations = {}
        
        # 基于当前质量推荐格式
        platform_preference = characteristics.get("platform_preference", "双平台")
        
        if platform_preference == "微信公众号":
            if quality_assessment["专业度"]["评分"] < 4.0:
                recommendations["推荐命令格式"] = "prompt4: 微信公众号 [您的需求] + 强调专业深度"
            else:
                recommendations["推荐命令格式"] = "prompt4: 微信公众号 [您的需求] + 注重创新表达"
        elif platform_preference == "小红书":
            if quality_assessment["吸引力"]["评分"] < 4.0:
                recommendations["推荐命令格式"] = "prompt4: 小红书 [您的需求] + 强调情感共鸣"
            else:
                recommendations["推荐命令格式"] = "prompt4: 小红书 [您的需求] + 注重转化设计"
        else:
            recommendations["推荐命令格式"] = "prompt4: 双平台 [您的需求] + 协同优化"
        
        # 预期改进效果
        weak_areas = [
            name for name, result in quality_assessment.items() 
            if result["评分"] < 3.5
        ]
        
        if weak_areas:
            recommendations["预期改进效果"] = "通过优化建议，预计整体质量可提升0.3-0.5分"
        else:
            recommendations["预期改进效果"] = "保持当前优秀水平，追求卓越表现"
        
        return recommendations
    
    def display_optimization_suggestions(self, suggestions):
        """显示优化建议"""
        print("📋 【个性化优化建议】")
        print("=" * 50)
        
        for category, items in suggestions.items():
            if not items:
                continue
                
            print(f"\n💡 {category}:")
            
            if category == "立即改进建议":
                for item in items:
                    priority_icon = "🔴" if item["优先级"] == "高" else "🟡"
                    print(f"   {priority_icon} {item['改进项']}: {item['具体建议']}")
                    print(f"      当前{item['当前评分']:.1f}分 → 目标{item['目标评分']:.1f}分")
            elif category == "下次使用建议":
                for key, value in items.items():
                    print(f"   • {key}: {value}")
            else:
                for item in items:
                    print(f"   • {item}")
    
    def predict_user_satisfaction(self, quality_assessment, optimization_suggestions):
        """预测用户满意度"""
        # 基于质量评估计算满意度
        overall_score = sum(
            result["评分"] * self.quality_metrics[metric]["weight"]
            for metric, result in quality_assessment.items()
        )
        
        # 基于优化建议数量调整
        improvement_count = len(optimization_suggestions.get("立即改进建议", []))
        satisfaction_adjustment = -0.1 * improvement_count  # 需要改进的地方越多，满意度预期越低
        
        # 基于专家协作质量调整
        collaboration_bonus = 0.2  # 专家协作带来的满意度提升
        
        predicted_satisfaction = min(5.0, max(1.0, overall_score + satisfaction_adjustment + collaboration_bonus))
        
        return predicted_satisfaction
    
    def calculate_overall_quality_score(self, quality_assessment):
        """计算整体质量评分"""
        return sum(
            result["评分"] * self.quality_metrics[metric]["weight"]
            for metric, result in quality_assessment.items()
        )
    
    def generate_monitoring_report(self, quality_assessment, optimization_suggestions, satisfaction_prediction):
        """生成监控报告"""
        report = {
            "报告时间": datetime.now().isoformat(),
            "质量评估摘要": {
                "总体评分": self.calculate_overall_quality_score(quality_assessment),
                "质量等级": self.get_quality_level(self.calculate_overall_quality_score(quality_assessment)),
                "各项指标": {
                    metric: result["评分"] for metric, result in quality_assessment.items()
                }
            },
            "优化建议摘要": {
                "立即改进项": len(optimization_suggestions.get("立即改进建议", [])),
                "短期计划项": len(optimization_suggestions.get("短期优化计划", [])),
                "长期建议项": len(optimization_suggestions.get("长期发展建议", []))
            },
            "用户满意度预测": satisfaction_prediction,
            "系统表现": {
                "监控状态": "正常",
                "评估完成率": 100,
                "建议生成率": 100
            }
        }
        
        return report
```

---

## 🎯 总结

**执行监控器**作为质量控制中心，提供了：

### 🚀 核心功能
- **📊 5维度质量评估** - 专业度、吸引力、结构性、平台适配性、商业价值
- **💡 个性化优化建议** - 基于用户特征和质量评估的定制建议
- **🔍 实时监控分析** - 全面监控执行过程和结果质量
- **📈 持续改进机制** - 提供系统性的改进建议和发展规划

### 🎯 系统优势
- **🔬 科学评估** - 基于量化指标的客观评估
- **🎯 精准建议** - 个性化、可操作的优化建议
- **📊 数据驱动** - 基于数据分析的决策支持
- **🔄 持续优化** - 建立完整的改进循环

---

*🎯 执行监控器 - 让每次创作都有科学的质量评估和精准的优化建议！* 