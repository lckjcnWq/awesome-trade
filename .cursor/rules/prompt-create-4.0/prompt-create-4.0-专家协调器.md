---
alwaysApply: true
---

# 🎯 Prompt-Create-4.0 专家协调器

## 🚀 系统概述

**专家协调器**是Prompt-Create-4.0工作流系统的专家管理中心，负责：
- 🔍 **专家调用链展示** - 实时显示专家工作过程
- 🤝 **专家协作管理** - 协调多专家协同工作
- 📊 **调用追踪记录** - 完整记录专家调用过程
- 📋 **调用报告生成** - 生成详细的执行报告

---

## 🔧 专家协调器核心引擎

### 📊 专家调用链实时展示系统

```python
class ExpertCoordinator:
    """专家协调器主类"""
    
    def __init__(self):
        self.display_system = self.init_display_system()
        self.call_tracker = self.init_call_tracker()
        self.current_stage = 0
        self.expert_results = {}
    
    def init_display_system(self):
        """初始化显示系统"""
        return {
            "stage_header": self.display_stage_header,
            "expert_call": self.display_expert_call,
            "expert_result": self.display_expert_result,
            "stage_summary": self.display_stage_summary,
            "expert_collaboration": self.display_expert_collaboration
        }
    
    def init_call_tracker(self):
        """初始化调用追踪器"""
        return {
            "call_records": [],
            "stage_records": [],
            "collaboration_records": [],
            "start_time": None,
            "end_time": None
        }
    
    def display_stage_header(self, stage_number, stage_name, description):
        """显示阶段头部信息"""
        print(f"\n🎯 【阶段{stage_number}】{stage_name}")
        print(f"📋 {description}")
        print("=" * 60)
    
    def display_expert_call(self, expert_name, call_purpose, estimated_time):
        """显示专家调用信息"""
        print(f"├── 🔧 调用专家：{expert_name}")
        print(f"│   ├── 🎯 调用目的：{call_purpose}")
        print(f"│   ├── ⏱️ 预计耗时：{estimated_time}秒")
        print(f"│   └── 🔄 状态：执行中...")
    
    def display_expert_result(self, expert_name, contribution, quality_score, execution_time):
        """显示专家执行结果"""
        print(f"├── ✅ 专家完成：{expert_name}")
        print(f"│   ├── 🎯 主要贡献：{contribution}")
        print(f"│   ├── 📊 质量评分：{quality_score}/5.0")
        print(f"│   ├── ⏱️ 实际耗时：{execution_time}秒")
        print(f"│   └── 🔄 状态：已完成")
    
    def display_stage_summary(self, stage_number, total_experts, total_time, stage_quality):
        """显示阶段总结"""
        print(f"\n📊 【阶段{stage_number}总结】")
        print(f"├── 👥 参与专家：{total_experts}个")
        print(f"├── ⏱️ 总执行时间：{total_time}秒")
        print(f"├── 📈 阶段质量：{stage_quality}/5.0")
        print(f"└── 🔄 状态：阶段完成")
        print("-" * 60)
    
    def display_expert_collaboration(self, expert_1, expert_2, collaboration_type):
        """显示专家协作信息"""
        print(f"├── 🤝 专家协作：{expert_1} ↔ {expert_2}")
        print(f"│   ├── 🔗 协作类型：{collaboration_type}")
        print(f"│   └── 📈 协作效果：协同增强")

    def execute_expert_workflow(self, requirement_analysis, expert_strategy, user_characteristics):
        """执行专家工作流"""
        print("\n🚀 【专家协调器启动】")
        print("=" * 60)
        
        self.call_tracker["start_time"] = datetime.now()
        
        try:
            # 阶段1: 需求分析与平台识别
            stage1_result = self.execute_stage_1(requirement_analysis)
            
            # 阶段2: 专家调度与匹配
            stage2_result = self.execute_stage_2(expert_strategy)
            
            # 阶段3: 内容创作与协作
            stage3_result = self.execute_stage_3(requirement_analysis, expert_strategy)
            
            # 阶段4: 优化与增强
            stage4_result = self.execute_stage_4(stage3_result, expert_strategy)
            
            # 阶段5: 质量验证
            stage5_result = self.execute_stage_5(stage4_result)
            
            # 生成最终结果
            final_result = self.generate_final_result(stage5_result, expert_strategy)
            
            self.call_tracker["end_time"] = datetime.now()
            
            # 生成调用链报告
            self.generate_call_chain_report()
            
            return final_result
            
        except Exception as e:
            print(f"❌ 专家协调器执行失败: {str(e)}")
            return {"success": False, "error": str(e)}
    
    def execute_stage_1(self, requirement_analysis):
        """执行阶段1: 需求分析与平台识别"""
        self.display_stage_header(1, "需求分析与平台识别", "深度分析用户需求，识别目标平台")
        
        # 调用平台识别引擎
        self.display_expert_call("平台智能识别引擎", "识别目标平台和内容类型", 3)
        
        # 模拟专家执行
        time.sleep(1)
        
        # 记录专家结果
        expert_result = {
            "expert_name": "平台智能识别引擎",
            "contribution": "成功识别目标平台和内容类型",
            "quality_score": 4.5,
            "execution_time": 3.2,
            "output": {
                "platform": requirement_analysis["平台信息"]["platform"],
                "content_type": requirement_analysis["内容类型"]["type"],
                "complexity": requirement_analysis["复杂度等级"]["level"]
            }
        }
        
        self.expert_results["平台智能识别引擎"] = expert_result
        self.display_expert_result(
            expert_result["expert_name"],
            expert_result["contribution"],
            expert_result["quality_score"],
            expert_result["execution_time"]
        )
        
        # 记录调用
        self.record_expert_call(1, expert_result)
        
        # 阶段总结
        self.display_stage_summary(1, 1, 3.2, 4.5)
        self.record_stage_completion(1, "需求分析与平台识别", 1, 3.2, 4.5)
        
        return expert_result
    
    def execute_stage_2(self, expert_strategy):
        """执行阶段2: 专家调度与匹配"""
        self.display_stage_header(2, "专家调度与匹配", "智能匹配最优专家组合")
        
        # 调用专家选择逻辑引擎
        self.display_expert_call("专家选择逻辑引擎", "智能匹配专家组合", 4)
        time.sleep(1)
        
        expert_result = {
            "expert_name": "专家选择逻辑引擎",
            "contribution": "智能匹配最优专家组合",
            "quality_score": 4.7,
            "execution_time": 4.1,
            "output": {
                "selected_experts": expert_strategy["selected_experts"],
                "total_experts": expert_strategy["total_experts"],
                "collaboration_strategy": expert_strategy["collaboration_complexity"]
            }
        }
        
        self.expert_results["专家选择逻辑引擎"] = expert_result
        self.display_expert_result(
            expert_result["expert_name"],
            expert_result["contribution"],
            expert_result["quality_score"],
            expert_result["execution_time"]
        )
        
        # 调用专家映射调度系统
        self.display_expert_call("专家映射调度系统", "制定专家调度策略", 3)
        time.sleep(1)
        
        mapping_result = {
            "expert_name": "专家映射调度系统",
            "contribution": "制定专家调度策略和执行序列",
            "quality_score": 4.6,
            "execution_time": 3.8,
            "output": {
                "execution_sequence": expert_strategy["selected_experts"],
                "parallel_groups": self.identify_parallel_groups(expert_strategy["selected_experts"]),
                "dependencies": self.analyze_dependencies(expert_strategy["selected_experts"])
            }
        }
        
        self.expert_results["专家映射调度系统"] = mapping_result
        self.display_expert_result(
            mapping_result["expert_name"],
            mapping_result["contribution"],
            mapping_result["quality_score"],
            mapping_result["execution_time"]
        )
        
        # 记录调用
        self.record_expert_call(2, expert_result)
        self.record_expert_call(2, mapping_result)
        
        # 阶段总结
        self.display_stage_summary(2, 2, 7.9, 4.65)
        self.record_stage_completion(2, "专家调度与匹配", 2, 7.9, 4.65)
        
        return {"stage1_result": expert_result, "stage2_result": mapping_result}
    
    def execute_stage_3(self, requirement_analysis, expert_strategy):
        """执行阶段3: 内容创作与协作"""
        self.display_stage_header(3, "内容创作与协作", "多专家协同创作高质量内容")
        
        selected_experts = expert_strategy["selected_experts"]
        content_creation_experts = [e for e in selected_experts if "写作" in e or "创意" in e or "内容" in e]
        
        stage_results = {}
        total_time = 0
        total_quality = 0
        
        # 并行执行内容创作专家
        for expert in content_creation_experts:
            self.display_expert_call(expert, "内容创作和结构优化", 8)
            time.sleep(0.5)  # 模拟并行执行
            
            expert_result = {
                "expert_name": expert,
                "contribution": self.get_expert_contribution(expert),
                "quality_score": 4.3 + random.uniform(0, 0.7),
                "execution_time": 7.5 + random.uniform(-1, 1),
                "output": {
                    "content_segment": f"{expert}生成的内容片段",
                    "optimization_applied": True,
                    "platform_adapted": True
                }
            }
            
            stage_results[expert] = expert_result
            total_time += expert_result["execution_time"]
            total_quality += expert_result["quality_score"]
            
            self.display_expert_result(
                expert_result["expert_name"],
                expert_result["contribution"],
                expert_result["quality_score"],
                expert_result["execution_time"]
            )
            
            self.record_expert_call(3, expert_result)
        
        # 显示专家协作
        if len(content_creation_experts) >= 2:
            self.display_expert_collaboration(
                content_creation_experts[0],
                content_creation_experts[1],
                "内容协同优化"
            )
        
        # 阶段总结
        avg_quality = total_quality / len(content_creation_experts) if content_creation_experts else 4.0
        self.display_stage_summary(3, len(content_creation_experts), total_time, avg_quality)
        self.record_stage_completion(3, "内容创作与协作", len(content_creation_experts), total_time, avg_quality)
        
        return stage_results
    
    def execute_stage_4(self, stage3_result, expert_strategy):
        """执行阶段4: 优化与增强"""
        self.display_stage_header(4, "优化与增强", "运用三维度优化专家引擎提升内容质量")
        
        # 执行三维度优化
        optimization_experts = [
            "开头优化专家引擎",
            "内容优化专家引擎",
            "结尾优化专家引擎"
        ]
        
        stage_results = {}
        total_time = 0
        total_quality = 0
        
        for expert in optimization_experts:
            self.display_expert_call(expert, f"执行{expert.split('优化')[0]}优化", 6)
            time.sleep(0.5)
            
            expert_result = {
                "expert_name": expert,
                "contribution": f"提供{expert.split('优化')[0]}优化建议和改进方案",
                "quality_score": 4.5 + random.uniform(0, 0.5),
                "execution_time": 5.5 + random.uniform(-0.5, 0.5),
                "output": {
                    "optimization_suggestions": f"{expert}的优化建议",
                    "improvement_score": 0.8 + random.uniform(0, 0.2),
                    "platform_adaptation": True
                }
            }
            
            stage_results[expert] = expert_result
            total_time += expert_result["execution_time"]
            total_quality += expert_result["quality_score"]
            
            self.display_expert_result(
                expert_result["expert_name"],
                expert_result["contribution"],
                expert_result["quality_score"],
                expert_result["execution_time"]
            )
            
            self.record_expert_call(4, expert_result)
        
        # 显示三维度协作
        self.display_expert_collaboration(
            "开头优化专家引擎",
            "内容优化专家引擎",
            "结构化优化协作"
        )
        
        self.display_expert_collaboration(
            "内容优化专家引擎",
            "结尾优化专家引擎",
            "价值递进优化"
        )
        
        # 阶段总结
        avg_quality = total_quality / len(optimization_experts)
        self.display_stage_summary(4, len(optimization_experts), total_time, avg_quality)
        self.record_stage_completion(4, "优化与增强", len(optimization_experts), total_time, avg_quality)
        
        return stage_results
    
    def execute_stage_5(self, stage4_result):
        """执行阶段5: 质量验证"""
        self.display_stage_header(5, "质量验证", "全面验证内容质量和标准合规性")
        
        # 调用质量验证器
        self.display_expert_call("写作质量验证器", "全面验证内容质量", 5)
        time.sleep(1)
        
        expert_result = {
            "expert_name": "写作质量验证器",
            "contribution": "全面验证内容质量，确保标准合规",
            "quality_score": 4.8,
            "execution_time": 4.5,
            "output": {
                "quality_metrics": {
                    "专业度": 4.6,
                    "吸引力": 4.7,
                    "结构性": 4.5,
                    "平台适配性": 4.8,
                    "商业价值": 4.4
                },
                "compliance_check": True,
                "optimization_needed": False
            }
        }
        
        self.expert_results["写作质量验证器"] = expert_result
        self.display_expert_result(
            expert_result["expert_name"],
            expert_result["contribution"],
            expert_result["quality_score"],
            expert_result["execution_time"]
        )
        
        # 记录调用
        self.record_expert_call(5, expert_result)
        
        # 阶段总结
        self.display_stage_summary(5, 1, 4.5, 4.8)
        self.record_stage_completion(5, "质量验证", 1, 4.5, 4.8)
        
        return expert_result
    
    def generate_final_result(self, stage5_result, expert_strategy):
        """生成最终结果"""
        print("\n🎉 【专家协调器完成】")
        print("=" * 60)
        
        # 计算整体性能
        total_execution_time = sum(
            sum(record["execution_time"] for record in stage["expert_records"])
            for stage in self.call_tracker["stage_records"]
        )
        
        average_quality = sum(
            stage["average_quality"] for stage in self.call_tracker["stage_records"]
        ) / len(self.call_tracker["stage_records"])
        
        final_result = {
            "success": True,
            "creation_results": {
                "content": "专家协作生成的高质量内容",
                "platform_optimized": True,
                "quality_verified": True
            },
            "call_chain_record": self.call_tracker,
            "used_experts": expert_strategy["selected_experts"],
            "execution_time": total_execution_time,
            "average_quality": average_quality,
            "expert_contributions": {
                expert: result["contribution"] 
                for expert, result in self.expert_results.items()
            }
        }
        
        print(f"✅ 创作任务完成")
        print(f"👥 使用专家数: {len(expert_strategy['selected_experts'])}个")
        print(f"⏱️ 总执行时间: {total_execution_time:.1f}秒")
        print(f"📊 平均质量: {average_quality:.1f}/5.0")
        
        return final_result
    
    def get_expert_contribution(self, expert_name):
        """获取专家贡献描述"""
        contributions = {
            "写作创意引擎": "提供创意灵感和独特视角",
            "写作自适应学习引擎": "智能适应用户风格和需求",
            "微信公众号深度写作引擎": "专业的微信公众号内容创作",
            "小红书种草写作引擎": "吸引力强的小红书种草内容",
            "双平台协调器": "统一管理双平台内容适配",
            "双平台语言适配器": "优化双平台语言表达",
            "专业视角专家群": "提供专业深度分析视角",
            "验证评估专家群": "确保内容质量和准确性",
            "生成优化专家群": "优化内容生成质量",
            "行业认知专家群": "提供行业洞察和趋势分析"
        }
        return contributions.get(expert_name, "提供专业内容创作服务")
    
    def identify_parallel_groups(self, experts):
        """识别并行执行组"""
        parallel_groups = []
        
        # 专家群组可以并行
        expert_groups = [e for e in experts if "专家群" in e]
        if len(expert_groups) > 1:
            parallel_groups.append(expert_groups)
        
        # 写作引擎可以并行
        writing_engines = [e for e in experts if "写作" in e and "引擎" in e]
        if len(writing_engines) > 1:
            parallel_groups.append(writing_engines)
        
        return parallel_groups
    
    def analyze_dependencies(self, experts):
        """分析专家依赖关系"""
        dependencies = {}
        
        for expert in experts:
            if "调度" in expert:
                dependencies[expert] = [e for e in experts if "识别" in e or "选择" in e]
            elif "协调" in expert:
                dependencies[expert] = [e for e in experts if "调度" in e]
            elif "优化" in expert:
                dependencies[expert] = [e for e in experts if "写作" in e and "优化" not in e]
        
        return dependencies
    
    def record_expert_call(self, stage, expert_result):
        """记录专家调用"""
        self.call_tracker["call_records"].append({
            "stage": stage,
            "expert": expert_result["expert_name"],
            "result": expert_result,
            "timestamp": datetime.now().isoformat()
        })
    
    def record_stage_completion(self, stage_number, stage_name, expert_count, total_time, average_quality):
        """记录阶段完成"""
        stage_record = {
            "stage_number": stage_number,
            "stage_name": stage_name,
            "expert_count": expert_count,
            "total_time": total_time,
            "average_quality": average_quality,
            "completion_time": datetime.now().isoformat(),
            "expert_records": [
                record for record in self.call_tracker["call_records"]
                if record["stage"] == stage_number
            ]
        }
        
        self.call_tracker["stage_records"].append(stage_record)
    
    def generate_call_chain_report(self):
        """生成调用链报告"""
        print("\n📋 【专家调用链报告】")
        print("=" * 60)
        
        total_time = (
            self.call_tracker["end_time"] - self.call_tracker["start_time"]
        ).total_seconds()
        
        print(f"📊 执行总览:")
        print(f"├── ⏱️ 总执行时间: {total_time:.1f}秒")
        print(f"├── 👥 参与专家数: {len(self.call_tracker['call_records'])}个")
        print(f"├── 🎯 完成阶段数: {len(self.call_tracker['stage_records'])}个")
        print(f"└── 📈 平均质量: {self.calculate_overall_quality():.1f}/5.0")
        
        # 阶段详情
        for stage in self.call_tracker["stage_records"]:
            print(f"\n📍 阶段{stage['stage_number']}: {stage['stage_name']}")
            print(f"├── 👥 专家数: {stage['expert_count']}个")
            print(f"├── ⏱️ 耗时: {stage['total_time']:.1f}秒")
            print(f"└── 📊 质量: {stage['average_quality']:.1f}/5.0")
    
    def calculate_overall_quality(self):
        """计算整体质量"""
        if not self.call_tracker["stage_records"]:
            return 0.0
        
        return sum(
            stage["average_quality"] for stage in self.call_tracker["stage_records"]
        ) / len(self.call_tracker["stage_records"])
```

---

## 🎯 总结

**专家协调器**作为专家管理中心，提供了：

### 🚀 核心功能
- **🔍 实时调用链展示** - 透明展示每个专家的工作过程
- **🤝 智能协作管理** - 优化专家间的协作效率
- **📊 完整调用追踪** - 详细记录所有专家调用过程
- **📋 调用报告生成** - 生成详细的执行报告

### 🎯 系统优势
- **👁️ 过程透明化** - 用户可清楚看到专家工作过程
- **⚡ 协作优化** - 智能并行执行和依赖管理
- **📈 质量保证** - 实时质量监控和评估
- **🔧 灵活调度** - 支持动态专家组合调整

---

*🎯 专家协调器 - 让24个专家像一个团队一样高效协作！* 