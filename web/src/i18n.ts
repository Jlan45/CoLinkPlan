import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

const resources = {
    en: {
        translation: {
            layout: {
                dashboard: "Dashboard",
                nodes: "Nodes",
                logout: "Logout"
            },
            login: {
                title: "Welcome Back",
                subtitle: "Sign in to your Co-Link account to manage your AI proxies.",
                email: "Email Address",
                password: "Password",
                submit: "Sign In",
                noAccount: "Don't have an account?",
                register: "Register for access",
                error: "Failed to login"
            },
            register: {
                title: "Create API Gateway",
                subtitle: "Start pooling your local AI resources seamlessly.",
                email: "Email Address",
                password: "Password",
                submit: "Register Account",
                hasAccount: "Already have an account?",
                login: "Sign in",
                error: "Failed to register"
            },
            dashboard: {
                title: "My Proxy Dashboard",
                subtitle: "Manage your access keys and gateway deployments.",
                apiToken: "Server API Token",
                apiTokenDesc: "Use this as standard OpenAI Bearer token",
                clientToken: "Gateway Client Token",
                clientTokenDesc: "Identity token for your local daemons",
                clientConfig: "Client Configuration",
                clientConfigDesc: "Add your Client Token to config.yaml to map local models",
                apiTest: "API Request Test",
                apiTestDesc: "Consume API exactly like standard SDK"
            },
            nodes: {
                title: "Active Network Nodes",
                subtitle: "Live view of connected local gateways providing compute.",
                online: "Online",
                emptyTitle: "No active nodes",
                emptyDesc: "Connect a local client to see it here.",
                penalized: "Penalized (Cooling Down)",
                healthy: "Healthy & Ready",
                capacity: "Capacity",
                models: "Models Advertised"
            },
            home: {
                badge: "Distributed AI Compute Gateway",
                heroTitle: "Pool Local AI Power. Use It Everywhere.",
                heroSubtitle: "Co-Link lets you aggregate local GPU machines into a unified OpenAI-compatible API. Route requests, balance load, and scale — all with zero cloud lock-in.",
                ctaStart: "Get Started Free",
                ctaNodes: "View Live Nodes",
                featuresTitle: "Why Co-Link?",
                featuresSubtitle: "Built for developers who want control.",
                feat1Title: "OpenAI Compatible",
                feat1Desc: "Drop-in replacement. Use any OpenAI SDK — no code changes required.",
                feat2Title: "Zero Trust Auth",
                feat2Desc: "Token-based access control for both API consumers and gateway nodes.",
                feat3Title: "Distributed Nodes",
                feat3Desc: "Connect unlimited local GPU machines as compute nodes across the network.",
                quickStartTitle: "Quick Start",
                quickStartSubtitle: "Up and running in under 5 minutes.",
                step1Title: "Install the Client",
                step1Desc: "Install the Co-Link gateway client on your local machine.",
                step1Comment: "Or build from source",
                step2Title: "Register an Account",
                step2Desc: "Create an account to receive your API and client tokens.",
                step3Title: "Configure & Run",
                step3Desc: "Create a config.yaml, then run the client daemon to join the network."
            }
        }
    },
    zh: {
        translation: {
            layout: {
                dashboard: "个人主页",
                nodes: "活跃节点",
                logout: "登出"
            },
            login: {
                title: "欢迎回来",
                subtitle: "登录您的 Co-Link 账户来管理您的 AI 代理网关。",
                email: "电子邮箱地址",
                password: "密码",
                submit: "登录",
                noAccount: "还没有账户？",
                register: "注册账号",
                error: "登录失败"
            },
            register: {
                title: "创建 AI API 网关",
                subtitle: "无缝聚合您的本地 AI 算力资源。",
                email: "电子邮箱地址",
                password: "密码",
                submit: "注册账号",
                hasAccount: "已有账户？",
                login: "登录",
                error: "注册失败"
            },
            dashboard: {
                title: "网关大盘面板",
                subtitle: "分别获取接入Token和客户端共享接入token，并管理网关部署。",
                apiToken: "服务端接口令牌 (API Token)",
                apiTokenDesc: "可直接作为标准 OpenAI SDK 的 Bearer Token 使用",
                clientToken: "客户端接入令牌 (Client Token)",
                clientTokenDesc: "作为您本地各个节点进程的身份标识凭证",
                clientConfig: "客户端接入教程",
                clientConfigDesc: "请将您的 Client Token 添加至 config.yaml 以映射本地模型",
                apiTest: "API调用教程",
                apiTestDesc: "完全遵循 OpenAI 标准 SDK 的请求方式发起调用"
            },
            nodes: {
                title: "活跃网络节点",
                subtitle: "实时展示目前活跃连接的各客户端节点与其负载情况。",
                online: "在线",
                emptyTitle: "目前暂无活跃节点",
                emptyDesc: "在本地启动带有 Client Token 的客户端进程后，它将显示在这里。",
                penalized: "已受惩罚 (冷却等待中)",
                healthy: "健康可用 (就绪)",
                capacity: "当前并发任务及上限",
                models: "挂载发布的本地模型"
            },
            home: {
                badge: "分布式 AI 算力代理网关",
                heroTitle: "聚合本地 AI 算力，统一对外提供服务",
                heroSubtitle: "Co-Link 让你将本地 GPU 机器聚合成一个统一的 OpenAI 兼容 API，实现请求路由、负载均衡和弹性伸缩 — 无需任何云平台锁定。",
                ctaStart: "免费开始使用",
                ctaNodes: "查看活跃节点",
                featuresTitle: "为什么选择 Co-Link？",
                featuresSubtitle: "为追求掌控权的开发者打造。",
                feat1Title: "OpenAI 完全兼容",
                feat1Desc: "即插即用的替代方案，无需修改任何代码，兼容所有 OpenAI SDK。",
                feat2Title: "零信任鉴权体系",
                feat2Desc: "基于令牌的访问控制，分别管理 API 调用者与网关节点的身份。",
                feat3Title: "分布式多节点",
                feat3Desc: "无限拓展，将任意数量的本地 GPU 机器接入全球计算网络。",
                quickStartTitle: "快速开始",
                quickStartSubtitle: "5 分钟内完成部署与接入。",
                step1Title: "安装客户端",
                step1Desc: "在你的本地机器上安装 Co-Link 网关客户端程序。",
                step1Comment: "或者从源码编译",
                step2Title: "注册账号",
                step2Desc: "注册账号以获取你的 API Token 和客户端接入 Token。",
                step3Title: "配置并运行",
                step3Desc: "创建 config.yaml 配置文件并启动客户端守护进程，加入算力网络。"
            }
        }
    }
};

i18n
    .use(LanguageDetector)
    .use(initReactI18next)
    .init({
        resources,
        fallbackLng: 'en',
        interpolation: {
            escapeValue: false
        }
    });

export default i18n;
