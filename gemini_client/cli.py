import argparse
import sys
import time
from .api import GeminiAPI


def main():
    parser = argparse.ArgumentParser(prog="gemini")
    parser.add_argument("--chat", type=str, help="Send message")
    parser.add_argument("--stream", action="store_true", help="Stream response")

    args = parser.parse_args()
    api = GeminiAPI()

    if args.chat:
        try:
            response = api.ask(args.chat, stream=False)
            if args.stream:
                for word in response.split():
                    print(word + " ", end="", flush=True)
                    time.sleep(0.05)
                print()
            else:
                print(response)
        except Exception as e:
            print(f"Error: {e}", file=sys.stderr)
            sys.exit(1)
    else:
        interactive(api)


def interactive(api):
    print("\n" + "=" * 60)
    print("Gemini Client")
    print("=" * 60)
    print("/stream - Toggle streaming")
    print("/clear  - Clear conversation")
    print("/quit   - Exit")
    print("=" * 60 + "\n")

    stream_mode = False

    while True:
        try:
            user_input = input("You: ").strip()
            if not user_input:
                continue

            if user_input.lower() == "/quit":
                print("Goodbye!")
                break
            elif user_input.lower() == "/stream":
                stream_mode = not stream_mode
                print(f"Streaming: {'ON' if stream_mode else 'OFF'}\n")
            elif user_input.lower() == "/clear":
                api.conversation_id = None
                api.response_id = None
                print("Cleared\n")
            else:
                try:
                    response = api.ask(user_input, stream=False)
                    print("\nGemini: ", end="", flush=True)
                    if stream_mode:
                        for word in response.split():
                            print(word + " ", end="", flush=True)
                            time.sleep(0.05)
                        print("\n")
                    else:
                        print(response + "\n")
                except Exception as e:
                    print(f"\nError: {e}\n")

        except (KeyboardInterrupt, EOFError):
            print("\nGoodbye!")
            break


if __name__ == "__main__":
    main()
