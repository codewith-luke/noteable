from pathlib import Path


def update_doc():
    # Path to the markdown file
    file_path = Path('./documents/test_document_3.md')  # Replace with the actual path to your file

    # Read the file
    with open(file_path, 'r') as file:
        lines = file.readlines()

    # Find the line with the dentist appointment and remove it
    dentist_line = None
    for i, line in enumerate(lines):
        if 'Dentist' in line:
            dentist_line = line
            del lines[i]

    # If the dentist line was found, add it to Wednesday's schedule
    if dentist_line:
        lines.append(dentist_line)

    # Write the updated lines back to the file
    with open(file_path, 'w') as file:
        file.writelines(lines)
