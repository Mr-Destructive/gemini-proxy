/**
 * Gemini API JavaScript Client
 * Free, anonymous access to Google Gemini API
 */

class GeminiAPI {
  constructor(options = {}) {
    this.baseUrl = "https://gemini.google.com";
    this.endpoint = "/_/BardChatUi/data/assistant.lamda.BardFrontendService/StreamGenerate";
    this.timeout = options.timeout || 60000;
    this.retries = options.retries || 3;
    this.conversationId = null;
    this.responseId = null;
  }

  _buildPayload(message) {
    const inner = [
      [message, 0, null, null, null, null, 0],
      ["en-US"],
      [this.conversationId, this.responseId] || [null, null],
    ];
    const outer = [null, JSON.stringify(inner)];
    return { "f.req": JSON.stringify(outer) };
  }

  _parseResponse(text) {
    if (text.startsWith(")]}'")) {
      text = text.slice(4);
    }

    try {
      const lines = text.trim().split("\n");
      const data = JSON.parse(lines[0]);

      if (Array.isArray(data)) {
        for (const item of data) {
          if (
            Array.isArray(item) &&
            item.length >= 3 &&
            item[0] === "wrb.fr" &&
            item[2]
          ) {
            try {
              const inner = JSON.parse(item[2]);
              if (
                Array.isArray(inner) &&
                inner.length > 4 &&
                inner[4]
              ) {
                for (const part of inner[4]) {
                  if (Array.isArray(part) && part.length > 1) {
                    let content = part[1];
                    if (Array.isArray(content) && content.length > 0) {
                      content = content[0];
                    }
                    if (typeof content === "string" && content.trim()) {
                      if (content.includes("```")) {
                        const segments = content.split("```");
                        if (segments.length >= 3) {
                          return segments[segments.length - 1].trim();
                        }
                      } else {
                        return content;
                      }
                    }
                  }
                }
              }
            } catch (e) {
              // continue
            }
          }
        }
      }
    } catch (e) {
      // continue
    }

    return "";
  }

  async ask(message, stream = false) {
    const url = this.baseUrl + this.endpoint;
    const payload = this._buildPayload(message);

    for (let attempt = 0; attempt < this.retries; attempt++) {
      try {
        const response = await this._fetchWithTimeout(url, {
          method: "POST",
          headers: {
            "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
            "Origin": this.baseUrl,
            "Referer": this.baseUrl + "/",
            "Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
          },
          body: new URLSearchParams(payload),
        });

        if (!response.ok) {
          throw new Error(`Status ${response.status}`);
        }

        const text = await response.text();
        const result = this._parseResponse(text);

        if (stream) {
          return this._streamGenerator(result);
        }
        return result;
      } catch (error) {
        if (
          (error.name === "AbortError" ||
            error.message.includes("timeout") ||
            error.message.includes("Failed to fetch")) &&
          attempt < this.retries - 1
        ) {
          const wait = (attempt + 1) * 2000;
          console.log(
            `Timeout, retrying in ${wait / 1000}s... (attempt ${
              attempt + 1
            }/${this.retries})`
          );
          await new Promise((resolve) => setTimeout(resolve, wait));
          continue;
        }
        throw new Error(
          `API error after ${this.retries} retries: ${error.message}`
        );
      }
    }
  }

  async _fetchWithTimeout(url, options) {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);

    try {
      return await fetch(url, {
        ...options,
        signal: controller.signal,
      });
    } finally {
      clearTimeout(timeoutId);
    }
  }

  async *_streamGenerator(text) {
    for (const word of text.split(" ")) {
      yield word + " ";
      await new Promise((resolve) => setTimeout(resolve, 50));
    }
  }

  clearConversation() {
    this.conversationId = null;
    this.responseId = null;
  }
}

// Export for different module systems
if (typeof module !== "undefined" && module.exports) {
  module.exports = GeminiAPI;
}
