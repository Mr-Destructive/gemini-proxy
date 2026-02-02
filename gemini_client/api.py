import json
import requests
from typing import Union, Generator


class GeminiAPI:
    BASE_URL = "https://gemini.google.com"
    API_ENDPOINT = "/_/BardChatUi/data/assistant.lamda.BardFrontendService/StreamGenerate"

    def __init__(self):
        self.session = requests.Session()
        self.session.headers.update({
            "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
            "Origin": self.BASE_URL,
            "Referer": f"{self.BASE_URL}/",
            "Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
        })
        self.conversation_id = None
        self.response_id = None

    def _build_payload(self, message: str) -> dict:
        inner = [
            [message, 0, None, None, None, None, 0],
            ["en-US"],
            [self.conversation_id, self.response_id] if self.conversation_id else [None, None],
        ]
        outer = [None, json.dumps(inner)]
        return {"f.req": json.dumps(outer)}

    def _parse_response(self, text: str) -> str:
        if text.startswith(")]}'"):
            text = text[4:]

        try:
            lines = text.strip().split('\n')
            data = json.loads(lines[0])

            if isinstance(data, list):
                for item in data:
                    if isinstance(item, list) and len(item) >= 3 and item[0] == "wrb.fr" and item[2]:
                        try:
                            inner = json.loads(item[2])
                            if isinstance(inner, list) and len(inner) > 4 and inner[4]:
                                for part in inner[4]:
                                    if isinstance(part, list) and len(part) > 1:
                                        text_content = part[1]
                                        if isinstance(text_content, list) and len(text_content) > 0:
                                            text_content = text_content[0]

                                        if isinstance(text_content, str) and text_content.strip():
                                            if "```" in text_content:
                                                segments = text_content.split("```")
                                                if len(segments) >= 3:
                                                    return segments[-1].strip()
                                            else:
                                                return text_content
                        except:
                            pass
        except:
            pass

        return ""

    def ask(self, message: str, stream: bool = False) -> Union[str, Generator]:
        url = f"{self.BASE_URL}{self.API_ENDPOINT}"
        payload = self._build_payload(message)

        try:
            response = self.session.post(url, data=payload, timeout=30)
            if response.status_code != 200:
                raise Exception(f"Status {response.status_code}")

            result = self._parse_response(response.text)

            if stream:
                def word_gen():
                    for word in result.split():
                        yield word + " "
                return word_gen()
            return result

        except Exception as e:
            raise Exception(f"API error: {e}")


def query(message: str) -> str:
    api = GeminiAPI()
    return api.ask(message, stream=False)
