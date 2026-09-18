# Monkey Intermediate Language (MKIL) 技术规范

本中间语言 `IR` 融合了高级托管运行时的核心抽象与现代化编译器后端的寄存器流及静态单赋值 SSA 设计，采用 **Block Arguments（块参数）** 代替 `Phi` 节点，并针对 AOT 编译、硬件对齐和零拷贝加载进行了极致优化。

---

## 1. 体系结构概述 (Architecture Overview)

MKIL（Monkey中间语言）是一个**基于寄存器、强类型、静态单赋值（SSA）形式**的托管中间表示（IR）。

### 1.1 设计哲学

* **静态单赋值（SSA）形式**：所有虚拟寄存器（以 `%` 命名）均具备唯一性，每个寄存器仅能被赋值一次。
* **显式控制流图（CFG）与块参数**：代码由明确的基本块（Basic Blocks）组成。摒弃传统的 `phi` 节点，全面采用 **Block Arguments（块参数）** 形式传递控制流状态。
* **解耦元数据（Decoupled Metadata）**：指令执行流中仅包含类型与字段的拓扑令牌（Tokens），物理字符串与反射元数据存放在独立的边带（Sidecar）文件中，支持 AOT 编译期的按需深度裁剪（Trimming）。

---

## 2. 类型系统 (Type System)

MKIL 拥有分层的类型系统，兼顾了底层硬件映射的高效性与高层面向对象语言的语义完整性。

### 2.1 纯量与矢量类型 (Scalar & Vector Types)

* `i1`：布尔类型。
* `i8`, `i16`, `i32`, `i64`：有符号/无符号整数（符号由具体操作指令的语义决定）。
* `f32`, `f64`：IEEE 754 浮点数。
* `v4f32`：128位矢量，包含 4 个 32 位浮点数。
* `v2i64`：128位矢量，包含 2 个 64 位整数。

### 2.2 托管与聚合类型 (Managed & Aggregate Types)

* `ptr`：通用不透明指针，其具体宽度（32位或64位）在 AOT/JIT 阶段由目标架构决定。
* `class <ClassName>`：托管类引用。该类型对应的内存对象受垃圾回收器（GC）追踪。
* `struct <StructName>`：值类型结构体。可在栈上或作为其他聚合类型的内联成员分配。
* `union <UnionName>`：原生标签联合（Tagged Union / Sum Type），用于实现高效的模式匹配。
* `trait <TraitName>`：特征/接口定义。用于约束泛型多态的动态或静态行为。
* `ref <Type>`：托管跟踪指针。可指向栈或托管堆内部，GC 会在搬迁内存时自动更新。
  * **静态生命周期域规则**：`ref` 类型属于栈帧敏感类型，严禁逃逸出当前函数域，不可作为 `newobj` 的字段。若作为块参数（Block Arguments）传递，目标基本块的生命周期域必须小于或等于当前支配树（Dominator Tree）的父节点。

---

## 3. 内存与生存期模型 (Memory & Lifetime Model)

为了消除全量垃圾回收带来的非确定性停顿，MKIL 指令集通过显式修饰符将内存分为三个管理域：

* `tracked`：托管堆内存。由高并发分代 GC 负责管理，JIT/AOT 编译器会为其自动生成精确的 **GC 栈映射表（GC Maps）**。
* `untracked`：非托管内存。直接映射到原生物理堆（如 C 语言的 malloc/free 区域），GC 不参与追踪。
* `scoped(lifetime_id)`：生命周期受限内存。由编译器在 IL 层面标记其作用域边界，允许将复杂的聚合类型在栈上分配，出域时由硬件指针直接回收，实现零成本的生命周期管理。

---

## 4. 指令集标准 (Instruction Set Standard)

所有的 MKIL 指令均采用一致的赋值语义：  
`[目标寄存器] = [操作码] [数据类型] [操作数列表]`

### 4.1 数据流与控制流指令

#### `mov`

将立即数或寄存器的值复制到目标寄存器。

* **语法**：`%dst = mov <type> <src>`
* **示例**：`%r1 = mov i32 1024`

#### `br` / `brcond`

显式块跳转指令，使用块参数传递 SSA 状态。

* **语法**：
  * `br label %block_name ( %val1, %val2, ... )`
  * `brcond i1 %cond, label %true_lbl( %args... ), label %false_lbl( %args... )`
* **示例**：`br label %loop_check(%next_idx, %next_sum)`

### 4.2 算术与高性能矢量指令

#### `add` / `sub` / `mul` / `div` / `mod`

基础算术指令。

* **语法**：`%dst = add <type> %src1, %src2`
* **示例**：`%sum = add i32 %r1, %r2`

#### `vadd` / `vsub` / `vmul` / `vdiv` / `vneg` / `vabs` / `vfma` / `vfms` / `vadds`

矢量运算指令。

* **语法**：`%dst = vadd <vectype> %src1, %src2`
* **示例**：`%v_res = vadd v4f32 %v1, %v2`

### 4.3 对象模型的托管指令

#### `newobj`

在托管堆上分配并初始化类对象。

* **语法**：`%dst = newobj class <ClassName>(<args>)`
* **示例**：`%user = newobj class App.User(%i32_id)`

#### `field.load`

安全地读取托管/非托管类型的字段。

* **语法**：`%dst = field.load <field_type> class <ClassName>::<FieldName>, %obj`
* **示例**：`%age = field.load i32 class App.User::Age, %user`

#### `field.store` / `field.store_ref`

显式写屏障字段写入指令。

* `field.store`：写入标量或非托管数据（例如 `i32`, `f64`），AOT/JIT 编译器直接生成硬件存储指令，**无 GC 屏障开销**。
* `field.store_ref`：写入托管对象引用（`class` 类型），AOT/JIT 编译器扫描到此操作码时**强制自动内联 GC 写屏障（Write Barrier）**。
* **语法**：
  * `field.store <field_type> class <ClassName>::<FieldName>, %obj, %val`
  * `field.store_ref class <ClassName>::<FieldName>, %obj, %val`

### 4.4 现代化调用族指令 (Calling Interface)

#### `call.virtual`

标准的虚方法调用，通过运行时的虚函数表（vtable）分发。

* **语法**：`%dst = call.virtual <ret_type> class <ClassName>::<Method>(<args>) %obj`

#### `call.witness`

**泛型核心优化指令。** 针对满足特定 `trait` 约束的泛型调用，通过显式传入的**伴随表（Witness Table）**获取类型的元数据与函数指针。从而在 IL 层即可避免值类型的膨胀，实现多态共享代码。

* **语法**：`%dst = call.witness <ret_type> trait <TraitName>::<Method>(<args>) %witness_table, %obj`
* **示例**：`%valid = call.witness i1 trait System.IEquatable::Equals(%r1) %wit_ptr, %r0`

#### `call.native`

**外部 FFI 接口指令。** 用于直接调用外部非托管 C 语言/Rust ABI 的物理动态库函数。带有显式 `unmanaged` 修饰符，强制 JIT/AOT 编译器在机器码生成阶段处理寄存器重排（SysV/AAPCS 适配）并自动前置插入 GC 线程隔离屏障，防止非托管执行流死锁托管世界的垃圾回收器。

* **语法**：`%dst = call.native <ret_type> @MethodName(<args>) unmanaged`
* **示例**：`%bytes_written = call.native i32 @write(i32 %fd, ptr %buf_ptr, i64 %len) unmanaged`

#### `match.union`

**原生标签联合模式匹配指令。** 提供在 IL 层的多路高效跳转，替代复杂的 `if-else` 或跳转表语法糖。

* **语法**：`match.union <UnionType> %union_reg [ case <Tag1> label %lbl1, case <Tag2> label %lbl2, default label %default_lbl ]`

---

## 5. 汇编文本对比示例 (Assembly Specification Comparison)

### 5.1 C# 源代码

```csharp
public struct Point 
{ 
    public int X; 
    public int Y; 
}

public int Calculate(Point p, int factor) 
{
    int result = (p.X + p.Y) * factor;
    return result;
}
```

### 5.2 传统 MSIL 实现 (ECMA-335 Stack-based)

```il
.method public hidebysig instance int32 Calculate(valuetype Point p, int32 factor) cil managed
{
    .maxstack 2
    .locals init (int32 V_0)
    
    ldarga.s     p
    ldfld        int32 Point::X
    ldarga.s     p
    ldfld        int32 Point::Y
    add
    ldarg.2      
    mul
    stloc.0      
    ldloc.0
    ret
}
```

### 5.3 现代 MKIL 实现 (Register-based SSA with Block Arguments)

```il
define i32 @Calculate(struct Point* %p, i32 %factor) managed {
entry:
    ; %p 为结构体只读指针，%factor 为直接映射的虚拟寄存器
    ; 寄存器直接持有强类型状态，不依赖外部计算栈
    
    %x = field.load i32 struct Point::X, %p      ; 直接提取偏移字段，无压栈开销
    %y = field.load i32 struct Point::Y, %p      ; 高效读取字段 Y 并赋给新寄存器 %y

    %sum = add i32 %x, %y                        ; 标准 SSA 加法指令
    %result = mul i32 %sum, %factor               ; 标准 SSA 乘法指令
    
    ret i32 %result                              ; 显式返回目标寄存器，AOT 阶段可直达物理寄存器                    
}
```

---

## 6. 元数据边带文件 (.mkmeta) 二进制布局规范

本规范定义了 **MKIL 元数据边带文件（后缀通常为 `.mkmeta`）**的底层二进制结构。该文件与执行码（`.mkil`）物理分离，所有表结构均基于固定大小的索引槽位，以支持 O(1) 时间复杂度的随机访问。

**所有结构体采取 4/8 字节对齐，严格禁止出现非标宽度的内存字段以确保满足零复制（Zero-Copy）`mmap` 加载性能**。

### 6.1 全局文件结构

```txt
+-------------------------------------------------------+
| 1. 文件头 (Header) - 固定 64 字节                     |
+-------------------------------------------------------+
| 2. 段索引表 (Section Table) - 动态大小                |
+-------------------------------------------------------+
| 3. 字符串与符号池 (String & Blob Pool)                |
+-------------------------------------------------------+
| 4. 类型与成员定义表 (Type & Member Definitions)       |
+-------------------------------------------------------+
| 5. 泛型与伴随映射表 (Generic & Witness Maps)          |
+-------------------------------------------------------+
```

### 6.2 详细二进制结构设计

#### 6.2.1 文件头 (Header) - 固定 64 字节

| 偏移量 (Offset) | 类型 (Type) | 字段名称 (Field) | 描述 (Description) |
| :--- | :--- | :--- | :--- |
| `0x00` | `u32` | `MagicNumber` | 固定的文件魔数：`0x4154474E` (ASCII: `NGTA`) |
| `0x04` | `u16` | `MajorVersion` | 主版本号 (例如: `1`) |
| `0x06` | `u16` | `MinorVersion` | 次版本号 (例如: `0`) |
| `0x08` | `u32` | `Flags` | 文件属性标志位（如 `0x1` 表示已深度剪裁） |
| `0x0C` | `u32` | `SectionCount` | 段索引表中的段总数 |
| `0x10` | `u64` | `FileLength` | 包含当前头在内的整包文件总字节数 |
| `0x18` | `u64` | `SourceHash` | 对应 `.mkil` 执行码文件的 SHA-256 截断哈希（安全绑定） |
| `0x20` | `u64` | `Reserved` | 保留字段，前瞻性硬件对齐占位 |
| `0x28` | `u64` | `Reserved` | 保留字段，前瞻性硬件对齐占位 |
| `0x30` | `u32` | `CheckSum` | 文件头后所有数据的 CRC32 校验和 |

#### 6.2.2 段索引表 (Section Table)

```c
struct SectionDescriptor {
    u32 SectionID;    // 段标识符 (1:字符串池, 2:类型定义, 3:字段, 4:方法, 5:伴随表)
    u32 Offset;       // 自文件起始处的绝对字节偏移量 (4字节对齐)
    u32 Length;       // 该段的数据总长度（字节）
    u32 EntryCount;   // 该段包含的结构条目总数
};
```

#### 6.2.3 字符串与符号池 (String & Blob Pool)

* 连续 UTF-8 编码存储所有文本。每个字符串开头采用 **LEB128 格式** 编码其字节长度。
* 不使用 `\0` 截断，完全通过长度进行切片读取。所有其他元数据表通过**相对池起始处的字节偏移量（StringOffset）**来引用字符串。

#### 6.2.4 类型定义表 (Type Definition Table)

```c
struct TypeDefEntry {
    u32 NameOffset;       // 指向字符串池的类名偏移量
    u32 NamespaceOffset;  // 指向字符串池的命名空间偏移量
    u32 Flags;            // 类型特征标志（如：0x1=Class, 0x2=Struct, 0x4=Interface）
    u32 ParentTypeToken;  // 父类的 TypeToken 基索引（如无则为 0xFFFFFFFF）
    
    u32 FieldStartToken;  // 隶属于该类型的第一个字段在 FieldTable 中的索引
    u32 FieldCount;       // 该类型拥有的字段总数
    u32 MethodStartToken; // 隶属于该类型的第一个方法在 MethodTable 中的索引
    u32 MethodCount;      // 该类型拥拥有的方法总数
};
```

#### 6.2.5 字段与方法定义表 (Field & Method Definition Tables)

```c
struct FieldDefEntry {
    u32 NameOffset;       // 字段名称在字符串池的偏移
    u32 TypeFlags;        // 字段类型标记（对应基本纯量类型 ID 或包含嵌套 TypeToken）
    u32 FieldOffset;      // 【AOT核心优化】该字段在实例内存中的静态字节偏移量
    u32 Attributes;       // 访问修饰符（如 Public, Private）
};

struct MethodDefEntry {
    u32 NameOffset;       // 方法名称在字符串池的偏移量
    u32 RvaOrInterpret;   // 指向 .mkil 中该方法代码基本块的相对虚拟地址（RVA）。若该方法为 FFI 导入 (Flags 包含 METHOD_FLAG_FFI)，则此值重定向为指向 FfiImportDescriptor。
    u16 Flags;            // 方法属性（如 Static, Virtual, Abstract, 0x8000=METHOD_FLAG_FFI）
    u16 MaxStackRegs;     // 优化：该方法执行所需的最大虚拟寄存器窗口大小
    u32 ReturnTypeFlags;  // 返回值类型描述
    u32 SignatureBlobOff; // 指向方法参数完整签名流的符号池偏移量
    u64 Reserved;         // 保留对齐位，整个结构体完美达成 32 字节宽并进行 8 字节边界对齐
};

// FFI 原生外部库函数静态导入描述符
struct FfiImportDescriptor {
    u32 ModuleNameOffset; // 指向字符串池，如 "libc.so" 或 "kernel32.dll"
    u32 EntryPointOffset; // 指向字符串池的具体 C 原生函数符号名称，如 "open"
    u32 CallingConvention;// 目标平台调用约定（0:默认Cdecl, 1:Stdcall, 2:Fastcall）
    u32 Reserved;         // 4字节补白，确保 8 字节硬件对齐
};
```

#### 6.2.6 泛型伴随映射表 (Generic Witness Map)

```c
struct WitnessMapEntry {
    u32 ImplementorToken; // 实现该 trait 的具体 TypeToken（如 Struct Point）
    u32 TraitToken;       // 目标 Trait/Interface 的 TypeToken
    u32 MemoryStride;     // 【关键】该类型的运行时物理内存步长（字节大小）
    u32 MemoryAlignment;  // 内存对齐边界要求（如 4字节、8字节对齐）

    u32 VTableSlotCount;  // 伴随表中的函数指针槽位总数
    u32 VTableOffsetsOff; // 指向符号池内的一个 u32 数组，记录每个 trait 方法对应的 MethodToken 映射关系
    u64 Reserved;         // 预留的 8 字节对齐槽位
};
```

---

## 7. 执行码文件 (.mkil) 二进制编码与指令流规范

本规范定义了 **MKIL 执行码文件（后缀为 `.mkil`）** 的底层二进制编码标准。作为基于寄存器的静态单赋值（SSA）指令表示，该文件格式旨在通过变长指令编码消除寄存器开销，并在无需解压的情况下实现 O(N) 线性流式单遍（Single-pass）验证与极速编译。

### 7.1 全局文件结构

```txt
+-------------------------------------------------------+
| 1. 文件头 (Header) - 固定 32 字节                     |
+-------------------------------------------------------+
| 2. 方法执行入口索引表 (Method RVA Table)              |
+-------------------------------------------------------+
| 3. 二进制指令流段 (Bytecode Instruction Segment)       |
+-------------------------------------------------------+
```

#### 7.1.1 文件头 (Header) - 固定 32 字节

| 偏移量 | 类型 | 字段名称 | 描述 |
| :--- | :--- | :--- | :--- |
| `0x00` | `u32` | `MagicNumber` | 固定的执行码魔数：`0x4C49434E` (ASCII: `NCIL`) |
| `0x04` | `u16` | `TargetArch` | 目标架构优化提示（`0`:通用IR, `1`:x86_64, `2`:ARM64） |
| `0x06` | `u16` | `Reserved` | 保留对齐字节 |
| `0x08` | `u32` | `MethodCount` | 文件中包含的方法（函数）总数 |
| `0x0C` | `u32` | `InstructionBytes` | 指令段的总字节大小（必须为 4 字节的倍数） |
| `0x10` | `u64` | `MetaBindHash` | 对应 `.mkmeta` 元数据文件的 SHA-256 截断哈希 |
| `0x18` | `u32` | `CheckSum` | 自 `0x00` 至 `0x17` 的 CRC32 校验和 |
| `0x1C` | `u32` | `Reserved` | 尾部保留对齐空间 |

#### 7.1.2 方法执行入口索引表 (Method RVA Table)

每个条目长 16 字节，严格保持 4 字节边界对齐。

```c
struct MethodRvaEntry {
    u32 MethodIndex;      // 对应元数据中的 MethodToken 索引
    u32 BytecodeOffset;   // 该方法在二进制指令流段中的起始绝对字节偏移量（强制 4 字节对齐）
    u32 LocalRegCount;    // 该方法声明使用的虚拟寄存器总数（用于分配编译期栈帧）
    u32 CodeSize;         // 该方法指令流占用的总字节数（不足 4 字节部分末尾填充 0x00 补齐）
};
```

---

### 7.2 虚拟寄存器编码规范

指令中的每个寄存器操作数均根据其索引大小，采用**前缀自适应变长编码**以压缩体积：

* **1 字节紧凑模式（Index 0 - 127）**：最高位（Bit 7）为 `0`。低 7 位直接代表寄存器索引。
  * *二进制形式*：`0xxxxxxx`
  * *覆盖范围*：可覆盖 90% 以上的高频常规函数运算。
* **2 字节扩展模式（Index 128 - 16,511）**：前两个比特位（Bit 7-6）为 `10`。后续 14 位代表索引值。
  * *二进制形式*：`10xxxxxx xxxxxxxx`
* **5 字节极端模式（Index > 16,511）**：前两个比特位（Bit 7-6）为 `11`。后续 1 字节的低 6 位与紧跟的 4 字节组合成一个完整的 38 位超大寄存器空间。
  * *二进制形式*：`11000000 [32位无符号整数]`

---

### 7.3 类型掩码 (Type Mask - 8比特)

类型掩码显式指示当前指令处理的数据类别与位宽，支持解码器直接在硬件寄存器间快速分发。

| Bit 7 - Bit 5 (基础类别) | Bit 4 - Bit 0 (精确大小/修饰) |
| :--- | :--- |
| `000`: 整数 (Int) | `00001`: 1位(i1) \| `01000`: 8位(i8) \| `10000`: 32位(i32) \| `11000`: 64位(i64) |
| `001`: 浮点 (Float) | `10000`: 32位(f32) \| `11000`: 64位(f64) |
| `010`: 矢量 (Vector) | `00001`: v4f32 \| `00010`: v2i64 |
| `011`: 托管引用 (Object) | `00000`: class 引用 \| `00001`: struct 引用 \| `00010`: ptr 裸指针 |

---

### 7.4 核心指令二进制负载布局

*(注：以下布局中涉及的虚拟寄存器均以 1 字节紧凑模式为例)*

#### 7.4.1 `add` (三地址算术加法) - 固定 5 字节

* **汇编表达**：`%r3 = add i32 %r1, %r2`
* **二进制布局**：
  `[Opcode (0x10)] [Type Mask (0x10)] [Dst Reg %r3] [Src1 Reg %r1] [Src2 Reg %r2]`

#### 7.4.2 `field.load` (读取类/结构体字段) - 变长 6~9 字节

* **汇编表达**：`%r5 = field.load i32 class User::Age, %r2`
* **二进制布局**：
  `[Opcode (0x35)] [Type Mask (0x60)] [Dst Reg %r5] [Obj Reg %r2] [FieldToken (LEB128)]`

#### 7.4.3 `field.store` / `field.store_ref` (写入字段) - 变长 6~9 字节

* **`field.store` 示例（纯数据写入，无 GC 屏障开销）**
  * 汇编：`field.store i32 class User::Age, %r2, %r5`
  * 布局：`[Opcode (0x36)] [Type Mask (0x10)] [Obj Reg %r2] [Src Reg %r5] [FieldToken (LEB128)]`
* **`field.store_ref` 示例（托管引用写入，强制内联 GC 写屏障）**
  * 汇编：`field.store_ref class User::BestFriend, %r2, %r6`
  * 布局：`[Opcode (0x37)] [Type Mask (0x60)] [Obj Reg %r2] [Src Reg %r6] [FieldToken (LEB128)]`

#### 7.4.4 `call.virtual` (虚方法多态调用) - 变长 6~9 字节

* **汇编表达**：`%r4 = call.virtual i32 class User::GetId() %r2`
* **二进制布局**：
  `[Opcode (0x50)] [Dst Reg %r4] [Obj Reg %r2] [MethodToken (LEB128)]`

#### 7.4.5 `call.witness` (伴随表泛型调用) - 变长 7~12 字节

* **汇编表达**：`%r4 = call.witness i32 trait Trait::Method(%r1) %r2`
* **二进制布局**：
  `[Opcode (0x52)] [Dst Reg %r4] [WitReg %r2] [ArgReg %r1] [ArgCount (u8)] [MethodToken (LEB128)]`

#### 7.4.6 `call.native` (外部原生 FFI 调用) - 变长 7~12 字节

* **汇编表达**：`%r5 = call.native i32 @puts(%r1) unmanaged`
* **二进制布局**：
  `[Opcode (0x55)] [Dst Reg %r5] [FirstArgReg %r1] [ArgCount (u8)] [MethodToken (LEB128)]`
  * *注：MethodToken 指向元数据中的 MethodDefEntry，进而重定向到 FfiImportDescriptor 获取模块和符号名称。*

#### 7.4.7 `match.union` (和类型模式匹配跳转) - 动态大小

* **汇编表达**：`match.union %r0 [ case 0 label %l1, case 1 label %l2 ]`
* **二进制布局**：
  `[Opcode (0x61)] [Src Reg %r0] [MatchCases (u16)] [Case 0 Payload] [Case 1 Payload]`
  * *单个 Case Payload 内部结构*：`[Tag ID (LEB128)] + [目标基本块相对当前指令的相对偏移量 RVA (i32)]`

---

## 8. 原生异常处理机制 (Exception Handling) 二进制编码规范

MKIL 采用**静态表驱动（Table-driven）与显式基本块（Basic Block）控制流关联**的设计。这种设计保证了高效率的 AOT 展开（零成本成功路径），并理顺了异常接管流与 Block Arguments 类型机制之间的底层冲突。

### 8.1 异常处理描述表 (EH Exception Table)

在方法的指令二进制数据段结束后，若该方法包含异常保护，则必须紧跟 8 字节边界对齐的 EHT 结构。

#### 8.1.1 EHT 段头结构

```c
struct EhTableHeader {
    u32 ClauseCount;      // 当前方法包含的异常子句（Try-Catch-Finally）总数
    u32 Reserved;         // 边界对齐保留，固定填充 0x00000000
};
```

#### 8.1.2 异常子句条目 (EhClauseEntry) - 固定 24 字节

```c
struct EhClauseEntry {
    u32 Flags;            // 子句类型：0x1=Catch, 0x2=Finally, 0x4=Fault
    u32 TryStartRva;      // 被保护区（Try）的起始指令相对方法开头的字节偏移
    u32 TryLength;        // 被保护区的总字节长度
    u32 HandlerStartRva;  // 处理器（Catch/Finally）入口基本块的相对字节偏移。
                          // 架构限制规则：目标基本块作为异常拦截入口，严格禁止声明或携带任何 Block Arguments。
    u32 CatchTypeToken;   // 针对 Catch 子句为异常类在 .mkmeta 中的 TypeToken；Finally 子句时固定为 0xFFFFFFFF
};
```

---

### 8.2 异常专用字节码指令

#### 8.2.1 `catchobj` (捕获异常对象) - 固定 3 字节

从运行时异常子系统中弹出当前的异常托管实例，并将其注入 SSA 虚拟寄存器。必须作为异常捕获块（Catch Block）的起始指令。

* **语法**：`%dst = catchobj class <ExceptionType>`
* **二进制布局**：`[Opcode (0x70)] [Type Mask (0x60)] [Dst Reg %r0]`

#### 8.2.2 `throw` (抛出/重抛异常) - 固定 3 字节

终止当前基本块执行，将控制权移交异常回溯系统。操作数寄存器有效时抛出新对象，为零时代表在 catch 内重抛（Rethrow）。

* **语法**：`throw %exception_reg`
* **二进制布局**：`[Opcode (0x71)] [Type Mask (0x60)] [Src Reg %r0]`

#### 8.2.3 `endfinally` (退出 Finally 保护) - 固定 1 字节

通知运行时当前基本块的清理逻辑完毕，恢复先前的回溯流或跳转执行流。

* **语法**：`endfinally`
* **二进制布局**：`[0x72]`

---

### 8.3 异常控制流 (EH CFG) 示例

以下通过汇编伪代码直观展示表驱动 EH 在控制流图中的组织逻辑：

```il
define void @EhExample() managed {
entry:
    br label %try_block

try_block:
    ; --- [EhClause 0 & 1 保护区间开始] ---
    call void @DoWork()
    ; --- [EhClause 0 & 1 保护区间结束] ---
    br label %finally_block_normal

catch_block:
    ; 异常控制流强行注入的拦截点。严禁带块参数，上下文通过第一条指令在内部安全重建。
    %ex = catchobj class System.ArgumentException
    call void @Log(class System.ArgumentException %ex)
    br label %finally_block_normal

finally_block_normal:
    br label %finally_entry

finally_entry:
    ; --- [Finally 块执行入口] ---
    call void @Cleanup()
    endfinally
}
```

---

## 9. 编译器验证与通过性约束指南 (Validator Compliance)

为确保 MKIL 执行码在进入 AOT/JIT 阶段前的强类型安全和内存完整性，验证器（Validator）必须对指令流进行单遍静态拓扑扫描（Single-pass Verification），并强制通过以下几项核心审查规则：

### 9.1 SSA 唯一性与支配性检测 (Dominance Compliance)

* **严格单赋值**：同一个虚拟寄存器索引在单个方法定义中，有且仅能作为一条指令的 `Dst Reg` 出现。
* **支配树规则**：除基本块参数（Block Arguments）在外，任何指令使用的 `Src Reg`，其定义所在的块（Defining Block）必须在控制流图（CFG）的支配树上严格支配（Dominate）当前使用块。

### 9.2 块参数类型拓扑匹配 (Block Argument Unification)

* 当执行跳转指令 `br` 或 `brcond` 指向包含参数的基本块时，跳转负载中携带的操作数寄存器数量及类型，必须与目标块首部声明的参数列表达成严格的一一映射（Bit-perfect Matching）。

### 9.3 托管指针托管越界防逃逸规则 (Anti-Escape Analysis)

* 所有具备 `ref <Type>` 属性的托管跟踪指针（ByRef 指针），禁止向任何生存期域（Lifetime Scope）宽于当前方法栈帧的聚合体逃逸。
* **非法动作定义**：禁止将 `ref` 数据作为 `field.store` 到任何 `class` 的堆字段中；禁止将 `ref` 指针作为当前方法的返回值（`ret`）传出，除非附加了明确与形参绑定的生存期令牌修饰符。
