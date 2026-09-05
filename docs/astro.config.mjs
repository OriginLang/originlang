// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	integrations: [
		starlight({
			title: 'OriginLang',
			description:
				'OriginLang is a polyglot, plugin-based platform foundation for building extensible products and services.',
			favicon: '/favicon.svg',
			locales: {
				root: { label: 'English', lang: 'en' },
				zh: { label: '简体中文', lang: 'zh-CN' },
			},
			social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/OriginLang/originlang' }],
			sidebar: [
				{
					label: 'Start Here',
					translations: { zh: '开始' },
					items: [
						{
							label: 'Introduction',
							translations: { zh: '简介' },
							link: 'getting-started/introduction',
						},
						{
							label: 'Quickstart',
							translations: { zh: '快速开始' },
							link: 'getting-started/quickstart',
						},
						{
							label: 'Project Structure',
							translations: { zh: '项目结构' },
							link: 'getting-started/project-structure',
						},
					],
				},
				{
					label: 'Architecture',
					translations: { zh: '架构' },
					items: [
						{
							label: 'Layered Architecture',
							translations: { zh: '分层架构' },
							link: 'architecture/overview',
						},
						{
							label: 'Go Host Kernel',
							translations: { zh: 'Go 宿主内核' },
							link: 'architecture/host-kernel',
						},
						{
							label: 'Transports',
							translations: { zh: '传输层' },
							link: 'architecture/transports',
						},
					],
				},
				{
					label: 'Guides',
					translations: { zh: '指南' },
					items: [
						{
							label: 'Write a Plugin',
							translations: { zh: '编写插件' },
							link: 'guides/write-a-plugin',
						},
						{
							label: 'Run a Host',
							translations: { zh: '运行宿主' },
							link: 'guides/run-a-host',
						},
						{
							label: 'Build with Bazel',
							translations: { zh: '使用 Bazel 构建' },
							link: 'guides/build-with-bazel',
						},
					],
				},
				{
					label: 'Reference',
					translations: { zh: '参考' },
					items: [
						{
							label: 'RPC Protocol',
							translations: { zh: 'RPC 协议' },
							link: 'reference/rpc-protocol',
						},
						{
							label: 'Error Codes',
							translations: { zh: '错误码' },
							link: 'reference/error-codes',
						},
						{
							label: 'Plugin Lifecycle',
							translations: { zh: '插件生命周期' },
							link: 'reference/plugin-lifecycle',
						},
						{
							label: 'Plugin Manifest',
							translations: { zh: '插件清单' },
							link: 'reference/manifest',
						},
						{
							label: 'Extension Points',
							translations: { zh: '扩展点' },
							link: 'reference/extension-points',
						},
						{
							label: 'CLI',
							translations: { zh: '命令行工具' },
							link: 'reference/cli',
						},
					],
				},
				{
					label: 'Roadmap',
					translations: { zh: '路线图' },
					link: 'roadmap',
				},
			],
		}),
	],
});