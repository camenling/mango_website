#include "common.h"
#include "macho.h"
#include "bundle.h"
#include "openssl.h"
#include "timer.h"
#include "archive.h"

#ifdef _WIN32
#include "common_win32.h"
#endif

#define ZSIGN_VERSION "0.7"

const struct option options[] = {
	{"debug", no_argument, NULL, 'd'},
	{"force", no_argument, NULL, 'f'},
	{"verbose", no_argument, NULL, 'V'},
	{"adhoc", no_argument, NULL, 'a'},
	{"cert", required_argument, NULL, 'c'},
	{"pkey", required_argument, NULL, 'k'},
	{"prov", required_argument, NULL, 'm'},
	{"password", required_argument, NULL, 'p'},
	{"bundle_id", required_argument, NULL, 'b'},
	{"bundle_name", required_argument, NULL, 'n'},
	{"bundle_version", required_argument, NULL, 'r'},
	{"entitlements", required_argument, NULL, 'e'},
	{"output", required_argument, NULL, 'o'},
	{"zip_level", required_argument, NULL, 'z'},
	{"dylib", required_argument, NULL, 'l'},
	{"weak", no_argument, NULL, 'w'},
	{"temp_folder", required_argument, NULL, 't'},
	{"sha256_only", no_argument, NULL, '2'},
	{"install", no_argument, NULL, 'i'},
	{"quiet", no_argument, NULL, 'q'},
	{"help", no_argument, NULL, 'h'},
	{}
};


#include "zsign.h"


int _zsign_main(int argc, char* argv[])
{
	ZTimer atimer;
	ZTimer gtimer;

	bool bForce = false;
	bool bInstall = false;
	bool bWeakInject = false;
	bool bAdhoc = false;
	bool bSHA256Only = false;
	uint32_t uZipLevel = 0;

	string strCertFile;
	string strPKeyFile;
	string strProvFile;
	string strPassword;
	string strBundleId;
	string strBundleVersion;
	string strOutputFile;
	string strDisplayName;
	string strEntitleFile;
	vector<string> arrDylibFiles;
	string strTempFolder = ZFile::GetTempFolder();
	string fromIpaPath;

	int step = 2;
	for (int i = 1; i < argc; i += step) {

            char* option = argv[i];
            if (option[1] == 'f' || option[1] == 'w' || option[1] == 'a' || option[1] == 'q' || option[1] == '2') {
                step = 1; // 单字符选项 fwa2q
            } else {
                step = 2; // 长选项
            }

            if (strcmp(option, "-k") == 0) {

                strPKeyFile = argv[i+1];

            } else if (strcmp(option, "-c") == 0) {

                strCertFile = argv[i+1];

            } else if (strcmp(option, "-p") == 0) {

                strPassword = argv[i+1];

            } else if (strcmp(option, "-m") == 0) {

                strProvFile = argv[i+1];

            } else if (strcmp(option, "-o") == 0) {

                strOutputFile = argv[i+1];

            } else if (strcmp(option, "-v") == 0) {

    			printf("version: %s\n", ZSIGN_VERSION);
    			return 0;

            } else if (strcmp(option, "-b") == 0) {

                strBundleId = argv[i+1];

            } else if (strcmp(option, "-n") == 0) {

                strDisplayName = argv[i+1];


            } else if (strcmp(option, "-z") == 0) {

                uZipLevel = atoi(argv[i+1]);


            } else if (strcmp(option, "-i") == 0) {

    			bInstall = true;

            } else if (strcmp(option, "-l") == 0) {

    			arrDylibFiles.push_back(ZFile::GetFullPath(argv[i+1]));

            } else if (strcmp(option, "-e") == 0) {

                strEntitleFile = ZFile::GetFullPath(argv[i+1]);

            } else if (strcmp(option, "-t") == 0) {

                strTempFolder = ZFile::GetFullPath(argv[i+1]);

            }  else if (strcmp(option, "-u") == 0) {

				fromIpaPath = ZFile::GetFullPath(argv[i+1]);

			} else if (strcmp(option, "-f") == 0) {
			    bForce = true;
            } else if (strcmp(option, "-w") == 0) {

                bWeakInject = true;

            } else if (strcmp(option, "-a") == 0) {

                bAdhoc = true;

            } else if (strcmp(option, "-2") == 0) {

                bSHA256Only = true;

            } else if (strcmp(option, "-q") == 0) {

                ZLog::SetLogLever(ZLog::E_NONE);

            }
        }


	 if (!ZFile::IsFolder(strTempFolder.c_str())) {
	 	ZLog::ErrorV(">>> Invalid temp folder! %s\n", strTempFolder.c_str());
	 	return -1;
	 }

	string strPath = fromIpaPath;
	if (!ZFile::IsFileExists(strPath.c_str())) {
		ZLog::ErrorV(">>> Invalid path! %s\n", strPath.c_str());
		return -1;
	}

	 if (uZipLevel < 0 || uZipLevel > 9) {
	 	ZLog::ErrorV(">>> Invalid zip level! Please input 0 - 9.\n");
	 	return -1;
	 }

	bool bZipFile = ZFile::IsZipFile(strPath.c_str());
	if (!bZipFile && !ZFile::IsFolder(strPath.c_str())) { // macho file
		ZMachO* macho = new ZMachO();
		if (!macho->Init(strPath.c_str())) {
			ZLog::ErrorV(">>> Invalid mach-o file! %s\n", strPath.c_str());
			return -1;
		}

		if (!bAdhoc && arrDylibFiles.empty() && (strPKeyFile.empty() || strProvFile.empty())) {
			macho->PrintInfo();
			return 0;
		}

		ZSignAsset zsa;
		if (!zsa.Init(strCertFile, strPKeyFile, strProvFile, strEntitleFile, strPassword, bAdhoc, bSHA256Only, true)) {
			return -1;
		}

		if (!arrDylibFiles.empty()) {
			for (string dyLibFile : arrDylibFiles) {
				if (!macho->InjectDylib(bWeakInject, dyLibFile.c_str())) {
					return -1;
				}
			}
		}

		atimer.Reset();
		ZLog::PrintV(">>> Signing:\t%s %s\n", strPath.c_str(), (bAdhoc ? " (Ad-hoc)" : ""));
		string strInfoSHA1;
		string strInfoSHA256;
		string strCodeResourcesData;
		bool bRet = macho->Sign(&zsa, bForce, strBundleId, strInfoSHA1, strInfoSHA256, strCodeResourcesData);
		atimer.PrintResult(bRet, ">>> Signed %s!", bRet ? "OK" : "Failed");
		return bRet ? 0 : -1;
	}

	bool bTempOutputFile = false;
	if (strOutputFile.empty()) {
		if (bInstall) {
			bTempOutputFile = true;
			strOutputFile = ZFile::GetRealPathV("%s/zsign_temp_%llu.ipa", strTempFolder.c_str(), ZUtil::GetMicroSecond());
		} else if (bZipFile) {
			ZLog::ErrorV(">>> Use -o option to specify the output file.\n");
			return -1;
		}
	}

	//init
	ZSignAsset zsa;
	if (!zsa.Init(strCertFile, strPKeyFile, strProvFile, strEntitleFile, strPassword, bAdhoc, bSHA256Only, false)) {
		return -1;
	}

	//extract
	bool bTempFolder = false;
	bool bEnableCache = true;
	string strFolder = strPath;
	if (bZipFile) {
		bForce = true;
		bTempFolder = true;
		bEnableCache = false;
		strFolder = ZFile::GetRealPathV("%s/zsign_folder_%llu", strTempFolder.c_str(), atimer.Reset());
		ZLog::PrintV(">>> Unzip:\t%s (%s) -> %s ... \n", strPath.c_str(), ZFile::GetFileSizeString(strPath.c_str()).c_str(), strFolder.c_str());
		if (!Zip::Extract(strPath.c_str(), strFolder.c_str())) {
			ZLog::ErrorV(">>> Unzip failed!\n");
			return -1;
		}
		atimer.PrintResult(true, ">>> Unzip OK!");
	}

	//sign
	atimer.Reset();
	ZBundle bundle;
	bool bRet = bundle.SignFolder(&zsa, strFolder, strBundleId, strBundleVersion, strDisplayName, arrDylibFiles, bForce, bWeakInject, bEnableCache);
	atimer.PrintResult(bRet, ">>> Signed %s!", bRet ? "OK" : "Failed");

	//archive
	if (bRet && !strOutputFile.empty()) {
		size_t pos = bundle.m_strAppFolder.rfind("Payload");
		if (string::npos != pos && pos > 0) {
			atimer.Reset();
			ZLog::PrintV(">>> Archiving: \t%s ... \n", strOutputFile.c_str());
			string strBaseFolder = bundle.m_strAppFolder.substr(0, pos - 1);
			if (!Zip::Archive(strBaseFolder.c_str(), strOutputFile.c_str(), uZipLevel)) {
				ZLog::Error(">>> Archive failed!\n");
				bRet = false;
			} else {
				atimer.PrintResult(true, ">>> Archive OK! (%s)", ZFile::GetFileSizeString(strOutputFile.c_str()).c_str());
			}
		} else {
			ZLog::Error(">>> Can't find payload directory!\n");
			bRet = false;
		}
	}

	//install
	if (bRet && bInstall) {
		bRet = ZUtil::SystemExecV("ideviceinstaller -i  \"%s\"", strOutputFile.c_str());
	}

	//clean
	if (bTempFolder) {
		ZFile::RemoveFolder(strFolder.c_str());
	}

	if (bTempOutputFile) {
		ZFile::RemoveFile(strOutputFile.c_str());
	}

	gtimer.Print(">>> Done.");
	return bRet ? 0 : -1;
}


// zsign.cpp 实现
__attribute__((visibility("default")))
extern "C" int ZSign(const ZSignOptions* options) {
    std::vector<const char*> argv = {"zsign"};
    std::vector<std::string> temp_strings;  // 用于保存临时字符串

    // 必需参数校验
    if (!options || !options->ipa_path || !options->p12_path || !options->provision_path || !options->output_path) {
        return -1;
    }

    // 构建参数列表
    auto add_arg = [&](const char* opt, const char* val = nullptr) {
        argv.push_back(opt);
        if (val) argv.push_back(val);
    };

    // 基本参数
    add_arg("-k", options->p12_path);
    add_arg("-m", options->provision_path);
    add_arg("-o", options->output_path);

    // 开关参数
//    if (options->adhoc) add_arg("-a"); // Ad-hoc
//    if (options->debug) add_arg("-d"); // debug
    if (options->force) add_arg("-f"); // 强制
//    if (options->weak_inject) add_arg("-w"); // 弱注入
//    if (options->sha256_only) add_arg("-2"); // 仅SHA256
//    if (options->quiet) add_arg("-q"); // 推出
//    if (options->install) add_arg("-i"); // 安装
//    if (options->show_help) add_arg("-h"); // 帮助
//    if (options->show_version) add_arg("-v"); // 版本

    // 带值参数
    if (options->cert_path) add_arg("-c", options->cert_path);
    if (options->password) add_arg("-p", options->password);
    if (options->bundle_id) add_arg("-b", options->bundle_id);
    if (options->bundle_name) add_arg("-n", options->bundle_name);
    if (options->bundle_version) add_arg("-r", options->bundle_version);
    if (options->entitlements) add_arg("-e", options->entitlements);
    if (options->temp_folder) add_arg("-t", options->temp_folder);

    // 数值参数
    // 处理zip_level
    if (options->zip_level >= 0 && options->zip_level <= 9) {
        temp_strings.push_back(std::to_string(options->zip_level));
        add_arg("-z", temp_strings.back().c_str());
    } else {
        std::cerr << "Invalid zip level! Please input 0 - 9." << std::endl;
        return -1;
    }

    // 确保dylib_paths数组以NULL结尾
    for (int i = 0; i < options->dylib_count; ++i) {
        if (options->dylib_paths[i]) {
            add_arg("-l", options->dylib_paths[i]);
        }
    }

    // 输入文件
    add_arg("-u", options->ipa_path);
    // argv.push_back(options->ipa_path);

    // 打印调试信息，查看传入的参数
    std::cout << "Arguments passed to zsign:" << std::endl;
    for (const auto& arg : argv) {
        std::cout << arg << " ";
    }
    std::cout << std::endl;

    // 转换参数类型
    std::vector<char*> args;
    for (auto& arg : argv) {
        args.push_back(const_cast<char*>(arg));
    }

    return _zsign_main(args.size(), args.data());
}