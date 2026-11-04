// zsign.h
#pragma once
#ifdef __cplusplus
extern "C" {
#endif

typedef struct {
    // 基本参数
    const char* ipa_path;          // 输入文件 (必需)
    const char* p12_path;          // -k (必需)
    const char* provision_path;    // -m (必需)
    const char* output_path;       // -o

    // 可选参数
    const char* cert_path;         // -c
    const char* password;          // -p
    const char* bundle_id;         // -b
    const char* bundle_name;       // -n
    const char* bundle_version;    // -r
    const char* entitlements;      // -e
    const char* temp_folder;       // -t

    // 多值参数
    const char** dylib_paths;      // -l
    int dylib_count;

    // 开关参数
    int adhoc;                     // -a
    int debug;                     // -d
    int force;                     // -f
    int weak_inject;               // -w
    int sha256_only;               // -2
    int quiet;                     // -q
    int install;                   // -i
    int show_help;                 // -h
    int show_version;              // -v

    // 数值参数
    int zip_level;                 // -z (0-9)
} ZSignOptions;

__attribute__((visibility("default")))
int ZSign(const ZSignOptions* options);

#ifdef __cplusplus
}
#endif