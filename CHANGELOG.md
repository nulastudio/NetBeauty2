## Change Log

### v1.2.8
#### 修复
1. 修复没有指定`excludes`时所有文件都被忽略的问题
2. Fix #8

### v1.2.7
#### 新增
1. `BeautyExcludes`选项

#### 修复
1. 修复默认`libsDir`没有正确更改为`libraries`的问题
2. Fix #7


### v1.2.6
#### 新增
1. `--gittree`以及`GitTree`选项


### v1.2.5
#### 更新
1. 默认`libsDir`从`runtimes`改为`libraries`

#### 优化
1. 优化RID匹配以及主程序集识别


### v1.2.4
#### 修复
1. 修复主程序集文件名判断逻辑错误


### v1.2.3
#### 修复
1. 第一次使用时会报两个文件读取失败的错误


### v1.2.2
#### 修复
1. `BeautyDir`在以`\`结尾时命令行解析有误导致目录获取错误的问题
2. Fix #5


### v1.2.1
#### 新增
1. `setcdn <gitcdn>`、`getcdn`、`delcdn` 功能


### v1.2.0
#### 修复
1. 某处文件名使用错误的问题

#### 优化
1. ArtifactVersion已独立开来，每个artifact拥有自己的版本号


### v1.1.8
#### 新增
1. `--force`以及`ForceBeauty`选项

#### 修复
1. Nuget包`DisablePatch`选项无法正确计算的问题
2. Nuget包`BeautyLogLevel`选项`Info`错写成`Log`的问题

#### 更新
1. 类库xml文件跟随程序集移动

#### 优化
1. Nuget使用示例镜像库改为gitee镜像
2. 移除`BeautyDir`选项多余的路径兼容性拼接

### v1.1.7
修复当发布路径包含空格时无法正确识别的问题


### v1.1.6
#### 修复
1. 发布目录获取不对的问题


### v1.1.5
#### 修复
1. 发布目录获取不对的问题


### v1.1.4
#### 修复
1. 3.0编译无法正确解析发布目录的问题


### v1.1.3
#### 修复
1. 获取不到3.x版本号的问题


### v1.1.2
#### 新增
1. .NETCore Global Tool


### v1.1.1
#### 修复
1. 文件夹权限问题


### v1.1.0
#### 更新
1. 默认`BeautyLibsDir`从`libraries`改为`runtimes`


### v1.0.4
#### 优化
项目整理


### v1.0.3
#### 新增
1. fxr patch
2. 自定义Patch镜像地址

#### 修复
1. fix #1


### v1.0.2
#### 新增
1. `BeautyAfterTasks`属性


### v1.0.1
#### 修复
1. bug fix

### v1.0.0
