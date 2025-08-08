---
alwaysApply: true
---

# 🎯 Prompt-Create-4.0 核心调度器

## 🚀 系统概述

**核心调度器**是Prompt-Create-4.0工作流系统的中央大脑，负责：
- 🔍 **需求智能解析** - 深度理解用户需求
- 🎯 **专家智能匹配** - 精准调度最适合的专家组合
- 🚀 **流程控制管理** - 确保工作流程的有序执行
- 🛡️ **错误处理机制** - 提供完善的异常处理和恢复

---

## 🔧 核心调度引擎

### 📋 主入口函数

```python
def prompt_create_4_0_core_scheduler(user_input, user_profile=None):
    """
    Prompt-Create-4.0 核心调度器主入口
    """
    # 导入协调器和监控器
    from prompt_create_4_0_expert_coordinator import ExpertCoordinator
    from prompt_create_4_0_execution_monitor import ExecutionMonitor
    
    coordinator = ExpertCoordinator()
    monitor = ExecutionMonitor()
    
    try:
        print("🎯 【核心调度器启动】")
        print("=" * 60)
        
        # 🔍 需求智能解析
        requirement_analysis = deep_requirement_analysis(user_input)
        
        # 👤 用户画像分析
        user_characteristics = analyze_user_profile(user_profile)
        
        # 🎯 专家智能匹配
        expert_strategy = intelligent_expert_matching(requirement_analysis, user_characteristics)
        
        # 🚀 调度专家协调器
        coordinator_result = coordinator.execute_expert_workflow(
            requirement_analysis, expert_strategy, user_characteristics
        )
        
        # 🔬 调度执行监控器
        monitor_result = monitor.quality_assessment_and_optimization(
            coordinator_result, user_characteristics
        )
        
        # 🎉 结果整合
        final_result = integrate_final_results(coordinator_result, monitor_result, requirement_analysis)
        
        print("\n🎉 【核心调度器完成】")
        return final_result
        
    except Exception as e:
        return handle_system_error(e, user_input)

def deep_requirement_analysis(user_input):
    """深度需求分析"""
    print("🔍 【深度需求分析】")
    
    # 平台识别
    platform_info = detect_target_platform(user_input)
    
    # 内容类型识别
    content_type = detect_content_type(user_input)
    
    # 复杂度评估
    complexity = assess_complexity(user_input)
    
    analysis_result = {
        "原始需求": user_input,
        "平台信息": platform_info,
        "内容类型": content_type,
        "复杂度等级": complexity,
        "分析时间": datetime.now().isoformat()
    }
    
    print(f"├── 🎯 目标平台: {platform_info['platform']}")
    print(f"├── 📝 内容类型: {content_type['type']}")
    print(f"└── 📊 复杂度级别: {complexity['level']}")
    
    return analysis_result

def detect_target_platform(content):
    """检测目标平台"""
    platform_keywords = {
        "微信公众号": ["微信公众号", "微信", "公众号"],
        "小红书": ["小红书", "小红书平台", "红书"],
        "双平台": ["双平台", "两个平台", "微信和小红书"],
        "通用": ["通用", "不限平台"]
    }
    
    for platform, keywords in platform_keywords.items():
        if any(keyword in content.lower() for keyword in keywords):
            return {"platform": platform, "confidence": 0.9}
    
    return {"platform": "双平台", "confidence": 0.5}

def detect_content_type(content):
    """检测内容类型"""
    type_patterns = {
        "深度分析": ["深度分析", "深度解析", "分析文章"],
        "实用教程": ["教程", "指南", "如何", "步骤"],
        "产品介绍": ["产品", "介绍", "推荐", "测评"],
        "营销文案": ["营销", "广告", "文案", "推广"],
        "知识科普": ["科普", "知识", "解释", "普及"]
    }
    
    for content_type, patterns in type_patterns.items():
        if any(pattern in content for pattern in patterns):
            return {"type": content_type, "confidence": 0.8}
    
    return {"type": "综合内容", "confidence": 0.4}

def assess_complexity(content):
    """评估复杂度"""
    complexity_score = (
        min(len(content) / 100, 3) +
        min(len(set(content.split())) / 20, 3) +
        (2 if "双平台" in content else 0)
    )
    
    if complexity_score <= 3:
        level = "简单"
    elif complexity_score <= 6:
        level = "中等"
    else:
        level = "复杂"
    
    return {"level": level, "score": complexity_score}

def analyze_user_profile(user_profile):
    """分析用户画像"""
    print("👤 【用户画像分析】")
    
    if not user_profile:
        user_profile = {}
    
    characteristics = {
        "experience_level": user_profile.get("experience_level", "中级"),
        "primary_goal": user_profile.get("primary_goal", "内容质量"),
        "platform_preference": user_profile.get("platform_preference", "双平台")
    }
    
    print(f"├── 🏆 用户等级: {characteristics['experience_level']}")
    print(f"├── 🎯 主要目标: {characteristics['primary_goal']}")
    print(f"└── 📱 平台偏好: {characteristics['platform_preference']}")
    
    return {"基础特征": characteristics}

def intelligent_expert_matching(requirement_analysis, user_characteristics):
    """智能专家匹配"""
    print("🎯 【智能专家匹配】")
    
    # 基础专家组合
    selected_experts = ["平台智能识别引擎", "专家选择逻辑引擎", "专家映射调度系统"]
    
    # 根据平台添加专家
    platform = requirement_analysis["平台信息"]["platform"]
    if platform == "微信公众号":
        selected_experts.extend(["微信公众号深度写作引擎", "专业视角专家群"])
    elif platform == "小红书":
        selected_experts.extend(["小红书种草写作引擎", "写作创意引擎"])
    elif platform == "双平台":
        selected_experts.extend(["双平台协调器", "双平台语言适配器"])
    
    # 根据复杂度添加专家
    if requirement_analysis["复杂度等级"]["level"] == "复杂":
        selected_experts.extend(["灵活专家组合器", "写作智能进化引擎"])
    
    # 添加优化专家引擎
    selected_experts.extend([
        "开头优化专家引擎",
        "内容优化专家引擎", 
        "结尾优化专家引擎"
    ])
    
    # 添加质量验证
    selected_experts.append("写作质量验证器")
    
    strategy = {
        "selected_experts": list(set(selected_experts)),
        "total_experts": len(set(selected_experts)),
        "collaboration_complexity": "中等" if len(selected_experts) <= 10 else "复杂"
    }
    
    print(f"├── 👥 选择专家数: {strategy['total_experts']}个")
    print(f"└── 🎪 协作复杂度: {strategy['collaboration_complexity']}")
    
    return strategy

def integrate_final_results(coordinator_result, monitor_result, requirement_analysis):
    """整合最终结果"""
    print("🔗 【结果整合】")
    
    final_result = {
        "success": True,
        "timestamp": datetime.now().isoformat(),
        "original_requirement": requirement_analysis["原始需求"],
        "platform_info": requirement_analysis["平台信息"],
        "creation_results": coordinator_result.get("creation_results", {}),
        "quality_assessment": monitor_result.get("quality_assessment", {}),
        "optimization_suggestions": monitor_result.get("optimization_suggestions", {}),
        "system_performance": {
            "total_experts_used": len(coordinator_result.get("used_experts", [])),
            "overall_quality_score": monitor_result.get("overall_quality_score", 0)
        }
    }
    
    print(f"├── ✅ 创作结果: 已生成")
    print(f"├── 📊 质量评估: {monitor_result.get('overall_quality_score', 0):.1f}/5.0")
    print(f"└── 💡 优化建议: 已生成")
    
    return final_result

def handle_system_error(error, user_input):
    """处理系统错误"""
    print(f"\n💥 【系统错误】: {str(error)}")
    
    return {
        "success": False,
        "error_type": "system_error",
        "error_message": str(error),
        "suggestion": "请稍后重试或联系技术支持"
    }
```

---

## 🎯 总结

**核心调度器**作为系统中央大脑，提供了：

### 🚀 核心功能
- **🔍 智能需求解析** - 深度理解用户需求和意图
- **🎯 专家智能匹配** - 基于需求特征的最优专家组合
- **🚀 流程精准调度** - 统一调度专家协调器和执行监控器
- **🛡️ 完善错误处理** - 多层次错误处理机制

### 🎯 系统优势
- **📊 数据驱动决策** - 基于量化分析的智能决策
- **🔧 灵活适应能力** - 支持不同复杂度和类型的需求
- **💪 高可靠性** - 完善的错误处理机制
- **🎪 协作优化** - 智能的专家协作管理

---

*🎯 核心调度器 - 让每个创作需求都能得到最智能、最精准的专家匹配和流程调度！* 