<h1 align="center">💦 BPB Wizard</h1>

A wizard to facilitate [BPB Panel](https://github.com/bia-pain-bache/BPB-Worker-Panel) deployments.

<p align="center">
  <img src="assets/wizard.jpg">
</p>
<br>

## 💡 Usage

> [!IMPORTANT]
> Please disconnect any Proxy or VPN before running wizard.

- You can download executable files from [Releases](https://github.com/bia-pain-bache/BPB-Wizard/releases) based on your OS, unzip and just run it.
- Android users (Termux) can use these scripts:

### ARM v8

```bash
curl -L -# -o BPB-Wizard.zip https://github.com/bia-pain-bache/BPB-Wizard/releases/latest/download/BPB-Wizard-linux-arm64.zip && unzip -o BPB-Wizard.zip && chmod +x ./BPB-Wizard-linux-arm64 && ./BPB-Wizard-linux-arm64
```

### ARM v7 (Old models)

```bash
curl -L -# -o BPB-Wizard.zip https://github.com/bia-pain-bache/BPB-Wizard/releases/latest/download/BPB-Wizard-linux-arm.zip && unzip -o BPB-Wizard.zip && chmod +x ./BPB-Wizard-linux-arm && ./BPB-Wizard-linux-arm
```

> [!TIP]
> 1- First it logs you in your Cloudflare account.
>
> 2- Wizard will ask some questions for setting Panel and Configs secrets. All secrets are generated safely and randomly. However, you can use default generated values or just enter desired values.
>
> 3- Opens your Panel in browser! Enjoy it...
