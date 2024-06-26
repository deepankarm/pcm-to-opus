import os
import sys
from typing import List


def read_all_pcm(directory: str) -> List[bytes]:
    pcm_files = [f for f in os.listdir(directory) if f.endswith(".pcm")]
    pcm_files.sort()
    pcm_data = []
    for pcm_file in pcm_files:
        with open(os.path.join(directory, pcm_file), "rb") as f:
            pcm_data.append(f.read())
    return pcm_data


def save_pcm_to_wave(pcm_data, output_file: str):
    import wave

    # Create a wave file
    with wave.open(output_file, "wb") as wf:
        wf.setnchannels(1)  # Mono
        wf.setsampwidth(2)  # Sample width in bytes (16-bit audio)
        wf.setframerate(24000)  # Sample rate
        for data in pcm_data:
            wf.writeframes(data)

    print(f"Saved PCM data to {output_file}")


if __name__ == "__main__":
    if len(sys.argv) != 3:
        print(f"Usage: {sys.argv[0]} <directory> <output_file>")
        sys.exit(1)

    directory = sys.argv[1]
    output_file = sys.argv[2]

    pcm_data = read_all_pcm(directory)
    save_pcm_to_wave(pcm_data, output_file)
